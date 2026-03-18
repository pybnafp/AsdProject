# Copyright (c) 2024 Microsoft Corporation.
# Licensed under the MIT License

"""All the steps to transform final relationships."""

import logging
from typing import Any
from uuid import uuid4

import pandas as pd

from graphrag.data_model.schemas import RELATIONSHIPS_FINAL_COLUMNS
from graphrag.index.operations.compute_degree import compute_degree
from graphrag.index.operations.compute_edge_combined_degree import (
    compute_edge_combined_degree,
)
from graphrag.index.operations.create_graph import create_graph

logger = logging.getLogger(__name__)




async def finalize_relationships(
    relationships: pd.DataFrame,
    entities: pd.DataFrame | None = None,
    config: Any = None,
    cache: Any = None,
    text_units: pd.DataFrame | None = None,
) -> pd.DataFrame:
    """All the steps to transform final relationships."""
    # 如果提供了实体数据，确保所有实体都有对应的关系
    if entities is not None:
        relationships = await ensure_all_entities_have_relationships(
            entities, relationships, config=config, cache=cache, text_units=text_units
        )
    

    graph = create_graph(relationships, edge_attr=["weight"])
    degrees = compute_degree(graph)

    # 按 (source, target) 去重
    final_relationships = relationships.drop_duplicates(subset=["source", "target"])
    final_relationships["combined_degree"] = compute_edge_combined_degree(
        final_relationships,
        degrees,
        node_name_column="title",
        node_degree_column="degree",
        edge_source_column="source",
        edge_target_column="target",
    )


    # 确保 text_unit_ids 列存在（兼容旧数据或数据转换过程中的问题）
    if "text_unit_ids" not in final_relationships.columns:
        final_relationships["text_unit_ids"] = [[] for _ in range(len(final_relationships))]

    final_relationships.reset_index(inplace=True)
    final_relationships["human_readable_id"] = final_relationships.index
    final_relationships["id"] = final_relationships["human_readable_id"].apply(
        lambda _x: str(uuid4())
    )

    return final_relationships.loc[
        :,
        RELATIONSHIPS_FINAL_COLUMNS,
    ]


async def ensure_all_entities_have_relationships(
    entities: pd.DataFrame,
    relationships: pd.DataFrame,
    config: Any = None,
    cache: Any = None,
    text_units: pd.DataFrame | None = None,
) -> pd.DataFrame:
    """确保所有实体都有对应的关系记录（为孤立实体创建新关系或自环关系）。"""
    # 获取所有实体名称
    entity_titles = set(entities["title"].dropna().unique())
    
    # 获取关系表中涉及的实体
    relationship_entities = set(relationships["source"].unique()) | set(relationships["target"].unique())
    
    # 找出孤立实体
    isolated_entities = entity_titles - relationship_entities
    
    if len(isolated_entities) == 0:
        return relationships
    
    # 如果有配置、缓存和文本单元，尝试为孤立实体创建新关系
    if config is not None and cache is not None and text_units is not None:
        from graphrag.index.operations.finalize_entities import create_relationships_for_isolated_entities
        
        try:
            new_relationships = await create_relationships_for_isolated_entities(
                isolated_entities, entities, text_units, config, cache
            )
            
            if new_relationships:
                new_relationships_df = pd.DataFrame(new_relationships)
                relationships = pd.concat([relationships, new_relationships_df], ignore_index=True)
                return relationships
        except Exception as e:
            logger.warning(f"Failed to create new relationships for isolated entities: {e}")
    
    # 回退到自环关系
    isolated_relationships = []
    for entity_title in isolated_entities:
        # 获取实体对应的text_unit_ids
        entity_row = entities[entities["title"] == entity_title]
        entity_text_unit_ids = []
        if not entity_row.empty:
            entity_text_unit_ids = entity_row.iloc[0].get("text_unit_ids", [])
        
        isolated_relationships.append({
            "source": entity_title,
            "target": entity_title,
            "weight": 0.1,  # 很低的权重
            "description": f"Self-reference for isolated entity: {entity_title}",
            "source_id": "system_generated",
            "text_unit_ids": entity_text_unit_ids,
            "combined_degree": 0,  # 初始值，后续会在finalize_relationships中重新计算
        })
    
    # 将孤立实体的关系添加到关系表中
    if isolated_relationships:
        isolated_df = pd.DataFrame(isolated_relationships)
        relationships = pd.concat([relationships, isolated_df], ignore_index=True)
    
    return relationships
