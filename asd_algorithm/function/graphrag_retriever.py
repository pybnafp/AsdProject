import pandas as pd
from typing import Any, Dict

from graphrag.config.embeddings import (
    entity_description_embedding,
    relationship_description_embedding,
)
from graphrag.config.models.graph_rag_config import GraphRagConfig
from graphrag.query.factory import get_local_search_engine, get_global_search_engine
from graphrag.query.indexer_adapters import (
    read_indexer_communities,
    read_indexer_covariates,
    read_indexer_entities,
    read_indexer_relationships,
    read_indexer_report_embeddings,
    read_indexer_reports,
    read_indexer_text_units,
)
from graphrag.utils.api import get_embedding_store, load_search_prompt


async def local_graph_search_only_context(
    config: GraphRagConfig,
    entities: pd.DataFrame,
    communities: pd.DataFrame,
    community_reports: pd.DataFrame,
    text_units: pd.DataFrame,
    relationships: pd.DataFrame,
    covariates: pd.DataFrame | None,
    community_level: int = 2,
    response_type: str = "Multiple Paragraphs",
    verbose: bool = False,
    query: str = "",
) -> Dict[str, Any]:
    """
    执行 GraphRAG 本地检索，仅返回图检索结果（不调用 LLM）。

    返回内容包括：
    - context_chunks: 拼接好的上下文文本（会被用作 system prompt 中的 {context_data}）
    - context_records: 各类图组件的明细（entities / relationships / sources / reports 等），已转换为 JSON 友好的结构
    """
    # 与 graphrag.api.query.local_search_streaming 保持一致的向量库与索引加载流程
    vector_store_args: Dict[str, Dict[str, Any]] = {}
    for index, store in config.vector_store.items():
        vector_store_args[index] = store.model_dump()

    # 实体/关系描述向量库
    entity_embedding_store = get_embedding_store(
        config_args=vector_store_args,
        embedding_name=entity_description_embedding,
    )
    relationship_embedding_store = get_embedding_store(
        config_args=vector_store_args,
        embedding_name=relationship_description_embedding,
    )

    # 索引数据适配为查询引擎的数据模型
    entities_ = read_indexer_entities(entities, communities, community_level)
    covariates_ = (
        read_indexer_covariates(covariates) if covariates is not None else []
    )
    reports = read_indexer_reports(
        community_reports,
        communities,
        community_level=community_level,
        dynamic_community_selection=False,
    )
    text_units_ = read_indexer_text_units(text_units)
    relationships_ = read_indexer_relationships(relationships)

    # local search 使用的 system prompt 模板（不直接传给 LLM，这里只为兼容 factory）
    prompt = load_search_prompt(config.root_dir, config.local_search.prompt)

    # 构建 local search 引擎，但不调用 stream_search / search
    search_engine = get_local_search_engine(
        config=config,
        reports=reports,
        text_units=text_units_,
        entities=entities_,
        relationships=relationships_,
        covariates={"claims": covariates_},
        response_type=response_type,
        entity_embedding_store=entity_embedding_store,
        relationship_embedding_store=relationship_embedding_store,
        system_prompt=prompt,
        callbacks=[],
    )

    # 只调用 context_builder.build_context，获取图检索结果
    context_result = search_engine.context_builder.build_context(
        query=query,
        **search_engine.context_builder_params,
    )

    # 将 DataFrame 转为 JSON 友好结构
    context_records_serialized: Dict[str, Any] = {}
    for key, value in (context_result.context_records or {}).items():
        if isinstance(value, pd.DataFrame):
            context_records_serialized[key] = value.to_dict(orient="records")
        else:
            context_records_serialized[key] = value

    return {
        "context_chunks": context_result.context_chunks,
        "context_records": context_records_serialized,
    }


async def global_graph_search_only_context(
    config: GraphRagConfig,
    entities: pd.DataFrame,
    communities: pd.DataFrame,
    community_reports: pd.DataFrame,
    community_level: int = 2,
    dynamic_community_selection: bool = False,
    response_type: str = "Multiple Paragraphs",
    verbose: bool = False,
    query: str = "",
) -> Dict[str, Any]:
    """
    执行 GraphRAG 全局检索，仅返回图检索结果（不调用 LLM）。

    返回内容包括：
    - context_chunks: 按批次划分的社区报告上下文文本列表（用于 map 阶段）
    - context_records: 全局检索使用到的社区报告等数据（JSON 友好结构）
    """
    # 与 graphrag.api.query.global_search_streaming 保持一致的数据适配流程
    communities_ = read_indexer_communities(communities, community_reports)
    reports = read_indexer_reports(
        community_reports,
        communities,
        community_level=community_level,
        dynamic_community_selection=dynamic_community_selection,
    )
    entities_ = read_indexer_entities(
        entities, communities, community_level=community_level
    )

    map_prompt = load_search_prompt(config.root_dir, config.global_search.map_prompt)
    reduce_prompt = load_search_prompt(
        config.root_dir, config.global_search.reduce_prompt
    )
    knowledge_prompt = load_search_prompt(
        config.root_dir, config.global_search.knowledge_prompt
    )

    search_engine = get_global_search_engine(
        config,
        reports=reports,
        entities=entities_,
        communities=communities_,
        response_type=response_type,
        dynamic_community_selection=dynamic_community_selection,
        map_system_prompt=map_prompt,
        reduce_system_prompt=reduce_prompt,
        general_knowledge_inclusion_prompt=knowledge_prompt,
        callbacks=[],
    )

    # 仅调用 context_builder.build_context，不运行 map/reduce LLM
    context_result = await search_engine.context_builder.build_context(
        query=query,
        conversation_history=None,
        **search_engine.context_builder_params,
    )

    context_records_serialized: Dict[str, Any] = {}
    for key, value in (context_result.context_records or {}).items():
        if isinstance(value, pd.DataFrame):
            context_records_serialized[key] = value.to_dict(orient="records")
        else:
            context_records_serialized[key] = value

    return {
        "context_chunks": context_result.context_chunks,
        "context_records": context_records_serialized,
    }

