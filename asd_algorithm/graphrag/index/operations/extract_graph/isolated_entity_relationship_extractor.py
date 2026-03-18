# Copyright (c) 2024 Microsoft Corporation.
# Licensed under the MIT License

"""孤立实体关系提取器"""

import logging
import re
import traceback
from typing import Any

import networkx as nx
import pandas as pd

from graphrag.config.defaults import graphrag_config_defaults
from graphrag.index.typing.error_handler import ErrorHandlerFn
from graphrag.index.utils.string import clean_str
from graphrag.language_model.protocol.base import ChatModel
from graphrag.prompts.index.extract_isolated_entity_relationships import (
    CONTINUE_ISOLATED_RELATIONSHIP_PROMPT,
    ISOLATED_ENTITY_RELATIONSHIP_EXTRACTION_PROMPT,
    LOOP_ISOLATED_RELATIONSHIP_PROMPT,
)

DEFAULT_TUPLE_DELIMITER = "<|>"
DEFAULT_RECORD_DELIMITER = "##"
DEFAULT_COMPLETION_DELIMITER = "<|COMPLETE|>"

logger = logging.getLogger(__name__)


class IsolatedEntityRelationshipExtractor:
    """专门用于为孤立实体提取关系的提取器"""

    def __init__(
        self,
        model_invoker: ChatModel,
        tuple_delimiter_key: str | None = None,
        record_delimiter_key: str | None = None,
        input_text_key: str | None = None,
        completion_delimiter_key: str | None = None,
        prompt: str | None = None,
        max_gleanings: int | None = None,
        on_error: ErrorHandlerFn | None = None,
    ):
        """初始化孤立实体关系提取器"""
        self._model = model_invoker
        self._input_text_key = input_text_key or "input_text"
        self._tuple_delimiter_key = tuple_delimiter_key or "tuple_delimiter"
        self._record_delimiter_key = record_delimiter_key or "record_delimiter"
        self._completion_delimiter_key = (
            completion_delimiter_key or "completion_delimiter"
        )
        self._extraction_prompt = prompt or ISOLATED_ENTITY_RELATIONSHIP_EXTRACTION_PROMPT
        self._max_gleanings = (
            max_gleanings
            if max_gleanings is not None
            else graphrag_config_defaults.extract_graph.max_gleanings
        )
        self._on_error = on_error or (lambda _e, _s, _d: None)

    async def extract_relationships_for_isolated_entity(
        self,
        isolated_entity: str,
        entity_description: str,
        text_content: str,
        existing_entities: set[str],
        candidate_entities: list[str] | None = None,
        entities: Any = None,
        prompt_variables: dict[str, Any] | None = None,
    ) -> list[dict[str, Any]]:
        """为孤立实体提取关系"""
        if prompt_variables is None:
            prompt_variables = {}

        # 准备提示词变量
        prompt_variables = {
            **prompt_variables,
            self._tuple_delimiter_key: prompt_variables.get(self._tuple_delimiter_key)
            or DEFAULT_TUPLE_DELIMITER,
            self._record_delimiter_key: prompt_variables.get(self._record_delimiter_key)
            or DEFAULT_RECORD_DELIMITER,
            self._completion_delimiter_key: prompt_variables.get(
                self._completion_delimiter_key
            )
            or DEFAULT_COMPLETION_DELIMITER,
            "isolated_entity": isolated_entity,
            "entity_description": entity_description,
            self._input_text_key: text_content,
            "candidate_entities": "\n".join(sorted(candidate_entities)) if candidate_entities else "",
        }

        try:
            # 调用LLM提取关系
            logger.info(
                "[Isolated-Relation] Start extraction: entity='%s', candidates=%d",
                isolated_entity,
                0 if not candidate_entities else len(candidate_entities),
            )
            result = await self._process_isolated_entity_document(
                isolated_entity, entity_description, text_content, prompt_variables
            )
            
            # 解析结果
            relationships = self._parse_relationship_results(
                result,
                prompt_variables.get(self._tuple_delimiter_key, DEFAULT_TUPLE_DELIMITER),
                prompt_variables.get(self._record_delimiter_key, DEFAULT_RECORD_DELIMITER),
                existing_entities,
                entities,
            )
            logger.info(
                "[Isolated-Relation] Parsed relationships for '%s': accepted=%d",
                isolated_entity,
                len(relationships),
            )
            
            return relationships
            
        except Exception as e:
            logger.exception("Error extracting relationships for isolated entity")
            self._on_error(
                e,
                traceback.format_exc(),
                {
                    "isolated_entity": isolated_entity,
                    "entity_description": entity_description,
                    "text_content": text_content[:200] + "..." if len(text_content) > 200 else text_content,
                },
            )
            return []

    async def _process_isolated_entity_document(
        self,
        isolated_entity: str,
        entity_description: str,
        text_content: str,
        prompt_variables: dict[str, str],
    ) -> str:
        """处理孤立实体的文档，提取关系"""
        response = await self._model.achat(
            self._extraction_prompt.format(**{
                **prompt_variables,
                "isolated_entity": isolated_entity,
                "entity_description": entity_description,
                self._input_text_key: text_content,
            }),
        )
        results = response.output.content or ""

        # 如果启用了多轮提取，继续提取更多关系
        if self._max_gleanings > 0:
            for i in range(self._max_gleanings):
                response = await self._model.achat(
                    CONTINUE_ISOLATED_RELATIONSHIP_PROMPT,
                    name=f"isolated-extract-continuation-{i}",
                    history=response.history,
                )
                results += response.output.content or ""

                # 如果是最后一轮，不需要检查是否继续
                if i >= self._max_gleanings - 1:
                    break

                response = await self._model.achat(
                    LOOP_ISOLATED_RELATIONSHIP_PROMPT,
                    name=f"isolated-extract-loopcheck-{i}",
                    history=response.history,
                )
                if response.output.content != "Y":
                    break

        return results

    def _parse_relationship_results(
        self,
        results: str,
        tuple_delimiter: str,
        record_delimiter: str,
        existing_entities: set[str],
        entities: Any = None,
    ) -> list[dict[str, Any]]:
        """解析关系提取结果"""
        relationships = []
        rejected = 0
        
        if not results.strip():
            return relationships

        records = [r.strip() for r in results.split(record_delimiter)]

        for record in records:
            record = re.sub(r"^\(|\)$", "", record.strip())
            record_attributes = record.split(tuple_delimiter)

            if (
                record_attributes[0] == '"relationship"'
                and len(record_attributes) >= 5
            ):
                source = clean_str(record_attributes[1].upper())
                target = clean_str(record_attributes[2].upper())
                relationship_description = clean_str(record_attributes[3])
                
                try:
                    weight = float(record_attributes[-1])
                except ValueError:
                    weight = 1.0

                # 验证关系中的实体是否存在于现有实体集合中
                if source in existing_entities and target in existing_entities:
                    # 获取target实体对应的text_unit_ids
                    target_text_unit_ids = []
                    if entities is not None:
                        target_entity_row = entities[entities["title"] == target]
                        if not target_entity_row.empty:
                            raw_text_unit_ids = target_entity_row.iloc[0].get("text_unit_ids", [])
                            # 处理可能的numpy数组或pandas Series
                            if hasattr(raw_text_unit_ids, '__len__') and len(raw_text_unit_ids) > 0:
                                target_text_unit_ids = list(raw_text_unit_ids)
                            elif raw_text_unit_ids:
                                target_text_unit_ids = [raw_text_unit_ids]
                    
                    relationships.append({
                        "source": source,
                        "target": target,
                        "weight": weight,
                        "description": relationship_description,
                        "source_id": "isolated_entity_extraction",
                        "text_unit_ids": target_text_unit_ids,
                        "combined_degree": 0,  # 初始值，后续会在finalize_relationships中重新计算
                    })
                else:
                    rejected += 1
                    logger.warning(
                        "Relationship contains non-existing entities: %s -> %s",
                        source,
                        target,
                    )

        if rejected > 0:
            logger.info("[Isolated-Relation] Rejected relationships due to non-existing entities: %d", rejected)
        return relationships
