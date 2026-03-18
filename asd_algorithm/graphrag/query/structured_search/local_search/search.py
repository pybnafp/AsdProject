# Copyright (c) 2024 Microsoft Corporation.
# Licensed under the MIT License

"""LocalSearch implementation."""

import logging
import time
from collections.abc import AsyncGenerator
from typing import Any

import tiktoken

from graphrag.callbacks.query_callbacks import QueryCallbacks
from graphrag.language_model.protocol.base import ChatModel
from graphrag.prompts.query.local_search_system_prompt import (
    LOCAL_SEARCH_SYSTEM_PROMPT,
)
from graphrag.query.context_builder.builders import LocalContextBuilder
from graphrag.query.context_builder.conversation_history import (
    ConversationHistory,
)
from graphrag.query.llm.text_utils import num_tokens
from graphrag.query.structured_search.base import BaseSearch, SearchResult

logger = logging.getLogger(__name__)
from pathlib import Path
from datetime import datetime
import pandas as pd


class LocalSearch(BaseSearch[LocalContextBuilder]):
    """Search orchestration for local search mode."""

    def __init__(
        self,
        model: ChatModel,
        context_builder: LocalContextBuilder,
        token_encoder: tiktoken.Encoding | None = None,
        system_prompt: str | None = None,
        response_type: str = "multiple paragraphs",
        callbacks: list[QueryCallbacks] | None = None,
        model_params: dict[str, Any] | None = None,
        context_builder_params: dict | None = None,
    ):
        super().__init__(
            model=model,
            context_builder=context_builder,
            token_encoder=token_encoder,
            model_params=model_params,
            context_builder_params=context_builder_params or {},
        )
        self.system_prompt = system_prompt or LOCAL_SEARCH_SYSTEM_PROMPT
        self.callbacks = callbacks or []
        self.response_type = response_type

    async def search(
        self,
        query: str,
        conversation_history: ConversationHistory | None = None,
        **kwargs,
    ) -> SearchResult:
        """Build local search context that fits a single context window and generate answer for the user query."""
        start_time = time.time()
        search_prompt = ""
        llm_calls, prompt_tokens, output_tokens = {}, {}, {}
        context_result = self.context_builder.build_context(
            query=query,
            conversation_history=conversation_history,
            **kwargs,
            **self.context_builder_params,
        )
        llm_calls["build_context"] = context_result.llm_calls
        prompt_tokens["build_context"] = context_result.prompt_tokens
        output_tokens["build_context"] = context_result.output_tokens

        logger.debug("GENERATE ANSWER: %s. QUERY: %s", start_time, query)
        print("context_result:", context_result)
        # 将query与所用entities/relationships输出到日志
        try:
            log_dir = Path("ragtest_test/logs")
            log_dir.mkdir(parents=True, exist_ok=True)
            log_file = log_dir / "query.log"

            records = context_result.context_records or {}
            entities_list: list[str] = []
            relationships_list: list[str] = []

            entities_df = records.get("entities")
            if isinstance(entities_df, pd.DataFrame) and not entities_df.empty:
                if "entity" in entities_df.columns:
                    entities_list = entities_df["entity"].astype(str).tolist()

            rels_df = records.get("relationships")
            if isinstance(rels_df, pd.DataFrame) and not rels_df.empty:
                src_col = "source" if "source" in rels_df.columns else None
                tgt_col = "target" if "target" in rels_df.columns else None
                if src_col and tgt_col:
                    relationships_list = [f"{str(s)} -> {str(t)}" for s, t in zip(rels_df[src_col], rels_df[tgt_col])]

            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            with log_file.open("a", encoding="utf-8") as f:
                f.write(f"[{timestamp}] mode=local\n")
                f.write(f"query: {query}\n")
                # 简要列表
                if entities_list:
                    f.write("entities: " + ", ".join(entities_list) + "\n")
                else:
                    f.write("entities: (none)\n")
                if relationships_list:
                    f.write("relationships: " + "; ".join(relationships_list) + "\n")
                else:
                    f.write("relationships: (none)\n")
                print("开始将entities/relationships详细记录输出到日志")
                # 详细记录：按上下文中实际使用到的数据行逐条输出
                # Entities 详细：human_readable_id(id), title(entity/title), type(可选), description
                if isinstance(entities_df, pd.DataFrame) and not entities_df.empty:
                    f.write("entities_detail:\n")
                    id_col = "id" if "id" in entities_df.columns else None
                    title_col = "entity" if "entity" in entities_df.columns else ("title" if "title" in entities_df.columns else None)
                    type_col = "type" if "type" in entities_df.columns else None
                    desc_col = "description" if "description" in entities_df.columns else None
                    for _, row in entities_df.iterrows():
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        etype = str(row[type_col]) if type_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"- id={hrid} | title={title} | type={etype} | description={desc}\n")
                else:
                    f.write("entities_detail: (none)\n")

                # Relationships 详细：human_readable_id(id), source, target, description
                if isinstance(rels_df, pd.DataFrame) and not rels_df.empty:
                    f.write("relationships_detail:\n")
                    id_col = "id" if "id" in rels_df.columns else None
                    src_col = "source" if "source" in rels_df.columns else None
                    tgt_col = "target" if "target" in rels_df.columns else None
                    desc_col = "description" if "description" in rels_df.columns else None
                    for _, row in rels_df.iterrows():
                        hrid = str(row[id_col]) if id_col else ""
                        src = str(row[src_col]) if src_col else ""
                        tgt = str(row[tgt_col]) if tgt_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"- id={hrid} | source={src} | target={tgt} | description={desc}\n")
                else:
                    f.write("relationships_detail: (none)\n")
                f.write("-" * 60 + "\n")
        except Exception:
            # 日志失败不影响主流程
            print("日志失败不影响主流程，未能将entities/relationships输出到日志")
            pass

        # 记录local search中使用的所有graph组件到专门的日志文件
        try:
            components_log_file = log_dir / "local_search_components.log"
            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            
            with components_log_file.open("a", encoding="utf-8") as f:
                f.write(f"[{timestamp}] LOCAL SEARCH COMPONENTS USAGE\n")
                f.write(f"Query: {query}\n")
                f.write("=" * 80 + "\n")
                
                # 记录所有可用的context_records
                records = context_result.context_records or {}
                f.write(f"Available context records: {list(records.keys())}\n\n")
                
                # 1. Entities 详细记录
                entities_df = records.get("entities")
                if isinstance(entities_df, pd.DataFrame) and not entities_df.empty:
                    f.write("ENTITIES USED:\n")
                    f.write(f"Total entities: {len(entities_df)}\n")
                    id_col = "id" if "id" in entities_df.columns else None
                    title_col = "entity" if "entity" in entities_df.columns else ("title" if "title" in entities_df.columns else None)
                    type_col = "type" if "type" in entities_df.columns else None
                    desc_col = "description" if "description" in entities_df.columns else None
                    for idx, (_, row) in enumerate(entities_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        etype = str(row[type_col]) if type_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"  {idx}. id={hrid} | title={title} | type={etype} | description={desc}\n")
                else:
                    f.write("ENTITIES USED: (none)\n")
                f.write("\n")
                
                # 2. Relationships 详细记录
                rels_df = records.get("relationships")
                if isinstance(rels_df, pd.DataFrame) and not rels_df.empty:
                    f.write("RELATIONSHIPS USED:\n")
                    f.write(f"Total relationships: {len(rels_df)}\n")
                    id_col = "id" if "id" in rels_df.columns else None
                    src_col = "source" if "source" in rels_df.columns else None
                    tgt_col = "target" if "target" in rels_df.columns else None
                    desc_col = "description" if "description" in rels_df.columns else None
                    for idx, (_, row) in enumerate(rels_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        src = str(row[src_col]) if src_col else ""
                        tgt = str(row[tgt_col]) if tgt_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"  {idx}. id={hrid} | source={src} | target={tgt} | description={desc}\n")
                else:
                    f.write("RELATIONSHIPS USED: (none)\n")
                f.write("\n")
                
                # 3. Text Units 详细记录
                text_units_df = records.get("sources")
                if isinstance(text_units_df, pd.DataFrame) and not text_units_df.empty:
                    f.write("TEXT UNITS USED:\n")
                    f.write(f"Total text units: {len(text_units_df)}\n")
                    print(f"text_units_df.head(): {text_units_df.head()}")
                    print(f"text_units_df.columns: {text_units_df.columns}")
                    id_col = "id" if "id" in text_units_df.columns else None
                    title_col = "title" if "title" in text_units_df.columns else None
                    content_col = "content" if "content" in text_units_df.columns else None
                    for idx, (_, row) in enumerate(text_units_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        content = str(row[content_col]) if content_col else ""
                        # 截断过长的内容
                        if len(content) > 200:
                            content = content[:200] + "..."
                        f.write(f"  {idx}. id={hrid} | title={title} | content={content}\n")
                else:
                    f.write("TEXT UNITS USED: (none)\n")
                f.write("\n")
                
                # 4. Communities 详细记录
                communities_df = records.get("reports")
                if isinstance(communities_df, pd.DataFrame) and not communities_df.empty:
                    f.write("COMMUNITIES USED:\n")
                    f.write(f"Total communities: {len(communities_df)}\n")
                    print(f"communities_df.head(): {communities_df.head()}")
                    print(f"communities_df.columns: {communities_df.columns}")
                    id_col = "id" if "id" in communities_df.columns else None
                    title_col = "title" if "title" in communities_df.columns else None
                    rank_col = "rank" if "rank" in communities_df.columns else None
                    summary_col = "summary" if "summary" in communities_df.columns else None
                    for idx, (_, row) in enumerate(communities_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        rank = str(row[rank_col]) if rank_col else ""
                        summary = str(row[summary_col]) if summary_col else ""
                        # 截断过长的摘要
                        if len(summary) > 300:
                            summary = summary[:300] + "..."
                        f.write(f"  {idx}. id={hrid} | title={title} | rank={rank} | summary={summary}\n")
                else:
                    f.write("COMMUNITIES USED: (none)\n")
                f.write("\n")
                
                # 5. Community Reports 详细记录（如果与communities不同）
                community_reports_df = records.get("community_reports")
                if isinstance(community_reports_df, pd.DataFrame) and not community_reports_df.empty:
                    f.write("COMMUNITY REPORTS USED:\n")
                    f.write(f"Total community reports: {len(community_reports_df)}\n")
                    id_col = "id" if "id" in community_reports_df.columns else None
                    title_col = "title" if "title" in community_reports_df.columns else None
                    rank_col = "rank" if "rank" in community_reports_df.columns else None
                    summary_col = "summary" if "summary" in community_reports_df.columns else None
                    for idx, (_, row) in enumerate(community_reports_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        rank = str(row[rank_col]) if rank_col else ""
                        summary = str(row[summary_col]) if summary_col else ""
                        # 截断过长的摘要
                        if len(summary) > 300:
                            summary = summary[:300] + "..."
                        f.write(f"  {idx}. id={hrid} | title={title} | rank={rank} | summary={summary}\n")
                else:
                    f.write("COMMUNITY REPORTS USED: (none)\n")
                f.write("\n")
                
                # 6. 其他可能的组件
                other_components = [key for key in records.keys() if key not in ["entities", "relationships", "sources", "reports", "community_reports"]]
                if other_components:
                    f.write("OTHER COMPONENTS USED:\n")
                    for component in other_components:
                        component_df = records.get(component)
                        if isinstance(component_df, pd.DataFrame) and not component_df.empty:
                            f.write(f"  {component}: {len(component_df)} records\n")
                        else:
                            f.write(f"  {component}: (none)\n")
                else:
                    f.write("OTHER COMPONENTS USED: (none)\n")
                
                f.write("\n" + "=" * 80 + "\n\n")
                
        except Exception as e:
            print(f"记录local search组件日志失败: {e}")
            pass
        try:
            if "drift_query" in kwargs:
                drift_query = kwargs["drift_query"]
                search_prompt = self.system_prompt.format(
                    context_data=context_result.context_chunks,
                    response_type=self.response_type,
                    global_query=drift_query,
                )
            else:
                search_prompt = self.system_prompt.format(
                    context_data=context_result.context_chunks,
                    response_type=self.response_type,
                )
            history_messages = [
                {"role": "system", "content": search_prompt},
            ]

            full_response = ""

            async for response in self.model.achat_stream(
                prompt=query,
                history=history_messages,
                model_parameters=self.model_params,
            ):
                full_response += response
                for callback in self.callbacks:
                    callback.on_llm_new_token(response)

            llm_calls["response"] = 1
            prompt_tokens["response"] = num_tokens(search_prompt, self.token_encoder)
            output_tokens["response"] = num_tokens(full_response, self.token_encoder)

            for callback in self.callbacks:
                callback.on_context(context_result.context_records)

            return SearchResult(
                response=full_response,
                context_data=context_result.context_records,
                context_text=context_result.context_chunks,
                completion_time=time.time() - start_time,
                llm_calls=sum(llm_calls.values()),
                prompt_tokens=sum(prompt_tokens.values()),
                output_tokens=sum(output_tokens.values()),
                llm_calls_categories=llm_calls,
                prompt_tokens_categories=prompt_tokens,
                output_tokens_categories=output_tokens,
            )

        except Exception:
            logger.exception("Exception in _asearch")
            return SearchResult(
                response="",
                context_data=context_result.context_records,
                context_text=context_result.context_chunks,
                completion_time=time.time() - start_time,
                llm_calls=1,
                prompt_tokens=num_tokens(search_prompt, self.token_encoder),
                output_tokens=0,
            )

    async def stream_search(
        self,
        query: str,
        conversation_history: ConversationHistory | None = None,
    ) -> AsyncGenerator:
        """Build local search context that fits a single context window and generate answer for the user query."""
        start_time = time.time()

        context_result = self.context_builder.build_context(
            query=query,
            conversation_history=conversation_history,
            **self.context_builder_params,
        )
        logger.debug("GENERATE ANSWER: %s. QUERY: %s", start_time, query)
        search_prompt = self.system_prompt.format(
            context_data=context_result.context_chunks, response_type=self.response_type
        )
        history_messages = [
            {"role": "system", "content": search_prompt},
        ]

        # 将query与所用entities/relationships输出到日志（stream 模式）
        try:
            log_dir = Path("ragtest_test/logs")
            log_dir.mkdir(parents=True, exist_ok=True)
            log_file = log_dir / "query.log"

            records = context_result.context_records or {}
            entities_df = records.get("entities")
            rels_df = records.get("relationships")

            entities_list: list[str] = []
            if isinstance(entities_df, pd.DataFrame) and not entities_df.empty:
                title_col = "entity" if "entity" in entities_df.columns else ("title" if "title" in entities_df.columns else None)
                if title_col:
                    entities_list = entities_df[title_col].astype(str).tolist()

            relationships_list: list[str] = []
            if isinstance(rels_df, pd.DataFrame) and not rels_df.empty:
                src_col = "source" if "source" in rels_df.columns else None
                tgt_col = "target" if "target" in rels_df.columns else None
                if src_col and tgt_col:
                    relationships_list = [f"{str(s)} -> {str(t)}" for s, t in zip(rels_df[src_col], rels_df[tgt_col])]

            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            with log_file.open("a", encoding="utf-8") as f:
                f.write(f"[{timestamp}] mode=local(stream)\n")
                f.write(f"query: {query}\n")
                if entities_list:
                    f.write("entities: " + ", ".join(entities_list) + "\n")
                else:
                    f.write("entities: (none)\n")
                if relationships_list:
                    f.write("relationships: " + "; ".join(relationships_list) + "\n")
                else:
                    f.write("relationships: (none)\n")

                # 详细记录
                if isinstance(entities_df, pd.DataFrame) and not entities_df.empty:
                    f.write("entities_detail:\n")
                    id_col = "id" if "id" in entities_df.columns else None
                    title_col = "entity" if "entity" in entities_df.columns else ("title" if "title" in entities_df.columns else None)
                    type_col = "type" if "type" in entities_df.columns else None
                    desc_col = "description" if "description" in entities_df.columns else None
                    for _, row in entities_df.iterrows():
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        etype = str(row[type_col]) if type_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"- id={hrid} | title={title} | type={etype} | description={desc}\n")
                else:
                    f.write("entities_detail: (none)\n")

                if isinstance(rels_df, pd.DataFrame) and not rels_df.empty:
                    f.write("relationships_detail:\n")
                    id_col = "id" if "id" in rels_df.columns else None
                    src_col = "source" if "source" in rels_df.columns else None
                    tgt_col = "target" if "target" in rels_df.columns else None
                    desc_col = "description" if "description" in rels_df.columns else None
                    for _, row in rels_df.iterrows():
                        hrid = str(row[id_col]) if id_col else ""
                        src = str(row[src_col]) if src_col else ""
                        tgt = str(row[tgt_col]) if tgt_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"- id={hrid} | source={src} | target={tgt} | description={desc}\n")
                else:
                    f.write("relationships_detail: (none)\n")
                f.write("-" * 60 + "\n")
        except Exception:
            pass

        # 记录local search stream模式中使用的所有graph组件到专门的日志文件
        try:
            components_log_file = log_dir / "local_search_components.log"
            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            
            with components_log_file.open("a", encoding="utf-8") as f:
                f.write(f"[{timestamp}] LOCAL SEARCH STREAM COMPONENTS USAGE\n")
                f.write(f"Query: {query}\n")
                f.write("=" * 80 + "\n")
                
                # 记录所有可用的context_records
                records = context_result.context_records or {}
                f.write(f"Available context records: {list(records.keys())}\n\n")
                
                # 1. Entities 详细记录
                entities_df = records.get("entities")
                if isinstance(entities_df, pd.DataFrame) and not entities_df.empty:
                    f.write("ENTITIES USED:\n")
                    f.write(f"Total entities: {len(entities_df)}\n")
                    id_col = "id" if "id" in entities_df.columns else None
                    title_col = "entity" if "entity" in entities_df.columns else ("title" if "title" in entities_df.columns else None)
                    type_col = "type" if "type" in entities_df.columns else None
                    desc_col = "description" if "description" in entities_df.columns else None
                    for idx, (_, row) in enumerate(entities_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        etype = str(row[type_col]) if type_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"  {idx}. id={hrid} | title={title} | type={etype} | description={desc}\n")
                else:
                    f.write("ENTITIES USED: (none)\n")
                f.write("\n")
                
                # 2. Relationships 详细记录
                rels_df = records.get("relationships")
                if isinstance(rels_df, pd.DataFrame) and not rels_df.empty:
                    f.write("RELATIONSHIPS USED:\n")
                    f.write(f"Total relationships: {len(rels_df)}\n")
                    id_col = "id" if "id" in rels_df.columns else None
                    src_col = "source" if "source" in rels_df.columns else None
                    tgt_col = "target" if "target" in rels_df.columns else None
                    desc_col = "description" if "description" in rels_df.columns else None
                    for idx, (_, row) in enumerate(rels_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        src = str(row[src_col]) if src_col else ""
                        tgt = str(row[tgt_col]) if tgt_col else ""
                        desc = str(row[desc_col]) if desc_col else ""
                        f.write(f"  {idx}. id={hrid} | source={src} | target={tgt} | description={desc}\n")
                else:
                    f.write("RELATIONSHIPS USED: (none)\n")
                f.write("\n")
                
                # 3. Text Units 详细记录
                text_units_df = records.get("sources")
                if isinstance(text_units_df, pd.DataFrame) and not text_units_df.empty:
                    f.write("TEXT UNITS USED:\n")
                    f.write(f"Total text units: {len(text_units_df)}\n")
                    print(f"text_units_df.head(): {text_units_df.head()}")
                    print(f"text_units_df.columns: {text_units_df.columns}")
                    id_col = "id" if "id" in text_units_df.columns else None
                    title_col = "title" if "title" in text_units_df.columns else None
                    text_col = "text" if "text" in text_units_df.columns else None
                    for idx, (_, row) in enumerate(text_units_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        content = str(row[text_col]) if text_col else ""
                        # 截断过长的内容
                        if len(content) > 200:
                            content = content[:200] + "..."
                        f.write(f"  {idx}. id={hrid} | title={title} | content={content}\n")
                else:
                    f.write("TEXT UNITS USED: (none)\n")
                f.write("\n")
                
                # 4. Communities 详细记录
                communities_df = records.get("reports")
                if isinstance(communities_df, pd.DataFrame) and not communities_df.empty:
                    f.write("COMMUNITIES USED:\n")
                    f.write(f"Total communities: {len(communities_df)}\n")
                    print(f"communities_df.head(): {communities_df.head()}")
                    print(f"communities_df.columns: {communities_df.columns}")
                    id_col = "id" if "id" in communities_df.columns else None
                    title_col = "title" if "title" in communities_df.columns else None
                    rank_col = "rank" if "rank" in communities_df.columns else None
                    summary_col = "summary" if "summary" in communities_df.columns else None
                    full_content_col = "full_content" if "full_content" in communities_df.columns else None
                    for idx, (_, row) in enumerate(communities_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        rank = str(row[rank_col]) if rank_col else ""
                        summary = str(row[summary_col]) if summary_col else ""
                        full_content = str(row[full_content_col]) if full_content_col else ""
                        # 截断过长的摘要
                        if len(summary) > 300:
                            summary = summary[:300] + "..."
                        if len(full_content) > 300:
                            full_content = full_content[:300] + "..."
                        f.write(f"  {idx}. id={hrid} | title={title} | rank={rank} | summary={summary} | full_content={full_content}\n")
                else:
                    f.write("COMMUNITIES USED: (none)\n")
                f.write("\n")
                
                # 5. Community Reports 详细记录（如果与communities不同）
                community_reports_df = records.get("community_reports")
                if isinstance(community_reports_df, pd.DataFrame) and not community_reports_df.empty:
                    f.write("COMMUNITY REPORTS USED:\n")
                    f.write(f"Total community reports: {len(community_reports_df)}\n")
                    id_col = "id" if "id" in community_reports_df.columns else None
                    title_col = "title" if "title" in community_reports_df.columns else None
                    rank_col = "rank" if "rank" in community_reports_df.columns else None
                    summary_col = "summary" if "summary" in community_reports_df.columns else None
                    for idx, (_, row) in enumerate(community_reports_df.iterrows(), 1):
                        hrid = str(row[id_col]) if id_col else ""
                        title = str(row[title_col]) if title_col else ""
                        rank = str(row[rank_col]) if rank_col else ""
                        summary = str(row[summary_col]) if summary_col else ""
                        # 截断过长的摘要
                        if len(summary) > 300:
                            summary = summary[:300] + "..."
                        f.write(f"  {idx}. id={hrid} | title={title} | rank={rank} | summary={summary}\n")
                else:
                    f.write("COMMUNITY REPORTS USED: (none)\n")
                f.write("\n")
                
                # 6. 其他可能的组件
                other_components = [key for key in records.keys() if key not in ["entities", "relationships", "sources", "reports", "community_reports"]]
                if other_components:
                    f.write("OTHER COMPONENTS USED:\n")
                    for component in other_components:
                        component_df = records.get(component)
                        if isinstance(component_df, pd.DataFrame) and not component_df.empty:
                            f.write(f"  {component}: {len(component_df)} records\n")
                        else:
                            f.write(f"  {component}: (none)\n")
                else:
                    f.write("OTHER COMPONENTS USED: (none)\n")
                
                f.write("\n" + "=" * 80 + "\n\n")
                
        except Exception as e:
            print(f"记录local search stream组件日志失败: {e}")
            pass

        for callback in self.callbacks:
            callback.on_context(context_result.context_records)

        async for response in self.model.achat_stream(
            prompt=query,
            history=history_messages,
            model_parameters=self.model_params,
        ):
            for callback in self.callbacks:
                callback.on_llm_new_token(response)
            yield response
