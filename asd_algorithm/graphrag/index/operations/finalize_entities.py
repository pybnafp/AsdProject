"""All the steps to transform final entities."""

import logging
from typing import Any
from uuid import uuid4

import pandas as pd

from graphrag.config.models.embed_graph_config import EmbedGraphConfig
from graphrag.data_model.schemas import ENTITIES_FINAL_COLUMNS
from graphrag.index.operations.compute_degree import compute_degree
from graphrag.index.operations.create_graph import create_graph
from graphrag.index.operations.embed_graph.embed_graph import embed_graph
from graphrag.index.operations.layout_graph.layout_graph import layout_graph

logger = logging.getLogger(__name__)


async def finalize_entities(
    entities: pd.DataFrame,
    relationships: pd.DataFrame,
    embed_config: EmbedGraphConfig | None = None,
    layout_enabled: bool = False,
    config: Any = None,
    cache: Any = None,
    text_units: pd.DataFrame | None = None,
) -> pd.DataFrame:
    """All the steps to transform final entities."""
    # 确保所有实体都参与图构建
    graph = await create_graph_with_all_entities(
        entities, relationships, edge_attr=["weight"], config=config, cache=cache, text_units=text_units
    )
    graph_embeddings = None
    if embed_config is not None and embed_config.enabled:
        graph_embeddings = embed_graph(
            graph,
            embed_config,
        )
    layout = layout_graph(
        graph,
        layout_enabled,
        embeddings=graph_embeddings,
    )
    degrees = compute_degree(graph)
    final_entities = (
        entities.merge(layout, left_on="title", right_on="label", how="left")
        .merge(degrees, on="title", how="left")
        .drop_duplicates(subset="title")
    )
    final_entities = final_entities.loc[entities["title"].notna()].reset_index()
    # disconnected nodes and those with no community even at level 0 can be missing degree
    final_entities["degree"] = final_entities["degree"].fillna(0).astype(int)
    final_entities.reset_index(inplace=True)
    final_entities["human_readable_id"] = final_entities.index
    final_entities["id"] = final_entities["human_readable_id"].apply(
        lambda _x: str(uuid4())
    )
    return final_entities.loc[
        :,
        ENTITIES_FINAL_COLUMNS,
    ]


async def create_graph_with_all_entities(
    entities: pd.DataFrame,
    relationships: pd.DataFrame,
    edge_attr: list[str | int] | None = None,
    config: Any = None,
    cache: Any = None,
    text_units: pd.DataFrame | None = None,
):
    """Create a networkx graph ensuring all entities are included as nodes."""
    import networkx as nx
    
    # 首先从关系创建基础图
    graph = nx.from_pandas_edgelist(relationships, edge_attr=edge_attr)
    
    # 获取所有实体名称
    entity_titles = set(entities["title"].dropna().unique())
    
    # 获取图中已有的节点
    existing_nodes = set(graph.nodes())
    
    # 为不在图中的实体添加节点
    isolated_entities = entity_titles - existing_nodes
    
    # 如果有配置、缓存和文本单元，尝试为孤立实体创建新关系
    if config is not None and cache is not None and text_units is not None and len(isolated_entities) > 0:
        new_relationships = await create_relationships_for_isolated_entities(
            isolated_entities, entities, text_units, config, cache
        )
        
        # 将新关系添加到图中
        for rel in new_relationships:
            graph.add_edge(
                rel["source"],
                rel["target"],
                weight=rel["weight"],
                description=rel["description"],
                source_id=rel["source_id"]
            )
    else:
        # 回退到自环关系
        for entity_title in isolated_entities:
            graph.add_edge(
                entity_title, 
                entity_title, 
                weight=0.1,  # 很低的权重表示自环
                description=f"Self-reference for isolated entity: {entity_title}",
                source_id="system_generated"
            )
    
    return graph


async def create_relationships_for_isolated_entities(
    isolated_entities: set[str],
    entities: pd.DataFrame,
    text_units: pd.DataFrame,
    config: Any,
    cache: Any,
) -> list[dict[str, Any]]:
    """为孤立实体创建新关系"""
    from graphrag.index.operations.extract_graph.isolated_entity_relationship_strategy import (
        create_isolated_entity_relationship_extractor,
    )
    
    # 获取所有现有实体名称（集合与有序列表分别用于校验和提示）
    existing_entities = set(entities["title"].dropna().unique())
    candidate_entities = sorted(existing_entities)
    
    all_new_relationships = []
    
    for entity_title in isolated_entities:
        # 获取实体的描述
        entity_row = entities[entities["title"] == entity_title]
        if entity_row.empty:
            continue
            
        entity_description = entity_row.iloc[0].get("description", "")
        
        # 获取相关的文本单元
        text_unit_ids = entity_row.iloc[0].get("text_unit_ids", [])
        # 处理可能的numpy数组或pandas Series
        if hasattr(text_unit_ids, '__len__') and len(text_unit_ids) == 0:
            continue
        elif hasattr(text_unit_ids, '__len__') and len(text_unit_ids) > 0:
            # 有内容，继续处理
            pass
        elif not text_unit_ids:
            continue
        else:
            # 其他情况，跳过
            continue
            
        # 收集相关文本
        related_texts = []
        for text_unit_id in text_unit_ids:
            text_unit_row = text_units[text_units["id"] == text_unit_id]
            if not text_unit_row.empty:
                related_texts.append(text_unit_row.iloc[0].get("text", ""))
        
        if not related_texts:
            continue
            
        # 合并文本内容
        combined_text = " ".join(related_texts)
        
        # 提取关系
        try:
            new_relationships = await create_isolated_entity_relationship_extractor(
                isolated_entity=entity_title,
                entity_description=entity_description,
                text_content=combined_text,
                existing_entities=existing_entities,
                candidate_entities=candidate_entities,
                entities=entities,
                cache=cache,
                args=config,
            )
            all_new_relationships.extend(new_relationships)
        except Exception as e:
            logger.warning(f"Failed to extract relationships for isolated entity {entity_title}: {e}")
            continue
    
    return all_new_relationships
