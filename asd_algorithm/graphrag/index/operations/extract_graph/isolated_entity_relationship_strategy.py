# Copyright (c) 2024 Microsoft Corporation.
# Licensed under the MIT License

"""孤立实体关系提取策略"""

import logging
from typing import Any

from graphrag.cache.pipeline_cache import PipelineCache
from graphrag.config.models.language_model_config import LanguageModelConfig
from graphrag.index.operations.extract_graph.isolated_entity_relationship_extractor import (
    IsolatedEntityRelationshipExtractor,
)
from graphrag.language_model.manager import ModelManager
from graphrag.language_model.protocol.base import ChatModel

logger = logging.getLogger(__name__)


async def run_isolated_entity_relationship_extraction(
    model: ChatModel,
    isolated_entity: str,
    entity_description: str,
    text_content: str,
    existing_entities: set[str],
    candidate_entities: list[str] | None,
    entities: Any,
    args: dict[str, Any],
) -> list[dict[str, Any]]:
    """运行孤立实体关系提取策略"""
    tuple_delimiter = args.get("tuple_delimiter", "<|>")
    record_delimiter = args.get("record_delimiter", "##")
    completion_delimiter = args.get("completion_delimiter", "<|COMPLETE|>")
    extraction_prompt = args.get("extraction_prompt", None)
    max_gleanings = args.get("max_gleanings", 1)

    extractor = IsolatedEntityRelationshipExtractor(
        model_invoker=model,
        prompt=extraction_prompt,
        max_gleanings=max_gleanings,
        on_error=lambda e, s, d: logger.error(
            "Isolated Entity Relationship Extraction Error", 
            exc_info=e, 
            extra={"stack": s, "details": d}
        ),
    )

    relationships = await extractor.extract_relationships_for_isolated_entity(
        isolated_entity=isolated_entity,
        entity_description=entity_description,
        text_content=text_content,
        existing_entities=existing_entities,
        candidate_entities=candidate_entities,
        entities=entities,
        prompt_variables={
            "tuple_delimiter": tuple_delimiter,
            "record_delimiter": record_delimiter,
            "completion_delimiter": completion_delimiter,
        },
    )

    return relationships


async def create_isolated_entity_relationship_extractor(
    isolated_entity: str,
    entity_description: str,
    text_content: str,
    existing_entities: set[str],
    candidate_entities: list[str] | None,
    entities: Any,
    cache: PipelineCache,
    args: dict[str, Any],
) -> list[dict[str, Any]]:
    """创建孤立实体关系提取器并运行提取"""
    llm_config = LanguageModelConfig(**args["llm"])

    llm = ModelManager().get_or_create_chat_model(
        name="extract_isolated_entity_relationships",
        model_type=llm_config.type,
        config=llm_config,
        cache=cache,
    )

    return await run_isolated_entity_relationship_extraction(
        model=llm,
        isolated_entity=isolated_entity,
        entity_description=entity_description,
        text_content=text_content,
        existing_entities=existing_entities,
        candidate_entities=candidate_entities,
        entities=entities,
        args=args,
    )
