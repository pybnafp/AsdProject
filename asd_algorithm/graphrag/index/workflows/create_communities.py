"""A module containing run_workflow method definition."""

import logging
from datetime import datetime, timezone
from typing import cast
from uuid import uuid4

import numpy as np
import pandas as pd

from graphrag.config.models.graph_rag_config import GraphRagConfig
from graphrag.data_model.schemas import COMMUNITIES_FINAL_COLUMNS
from graphrag.index.operations.cluster_graph import cluster_graph
from graphrag.index.operations.create_graph import create_graph
from graphrag.index.typing.context import PipelineRunContext
from graphrag.index.typing.workflow import WorkflowFunctionOutput
from graphrag.utils.storage import load_table_from_storage, write_table_to_storage

logger = logging.getLogger(__name__)


async def run_workflow(
    config: GraphRagConfig,
    context: PipelineRunContext,
) -> WorkflowFunctionOutput:
    """All the steps to transform final communities."""
    logger.info("Workflow started: create_communities")
    entities = await load_table_from_storage("entities", context.output_storage)
    relationships = await load_table_from_storage(
        "relationships", context.output_storage
    )

    max_cluster_size = config.cluster_graph.max_cluster_size
    use_lcc = config.cluster_graph.use_lcc
    seed = config.cluster_graph.seed

    output = create_communities(
        entities,
        relationships,
        max_cluster_size=max_cluster_size,
        use_lcc=use_lcc,
        seed=seed,
    )

    await write_table_to_storage(output, "communities", context.output_storage)

    logger.info("Workflow completed: create_communities")
    return WorkflowFunctionOutput(result=output)


def create_communities(
    entities: pd.DataFrame,
    relationships: pd.DataFrame,
    max_cluster_size: int,
    use_lcc: bool,
    seed: int | None = None,
) -> pd.DataFrame:
    """All the steps to transform final communities."""
    graph = create_graph(relationships, edge_attr=["weight"])

    clusters = cluster_graph(
        graph,
        max_cluster_size,
        use_lcc,
        seed=seed,
    )

    communities = pd.DataFrame(
        clusters, columns=pd.Index(["level", "community", "parent", "title"])
    ).explode("title")
    communities["community"] = communities["community"].astype(int)

    # aggregate entity ids for each community
    entity_ids = communities.merge(entities, on="title", how="inner")
    entity_ids = (
        entity_ids.groupby("community").agg(entity_ids=("id", list)).reset_index()
    )

    # aggregate relationships ids for each community
    # these are limited to only those where the source and target are in the same community
    max_level = communities["level"].max()
    all_grouped = pd.DataFrame(
        columns=["community", "level", "relationship_ids", "text_unit_ids"]  # type: ignore
    )
    for level in range(max_level + 1):
        communities_at_level = communities.loc[communities["level"] == level]
        sources = relationships.merge(
            communities_at_level, left_on="source", right_on="title", how="inner"
        )
        targets = sources.merge(
            communities_at_level, left_on="target", right_on="title", how="inner"
        )
        matched = targets.loc[targets["community_x"] == targets["community_y"]]
        
        # 聚合relationship_ids
        grouped = (
            matched.groupby(["community_x", "level_x", "parent_x"])
            .agg(
                relationship_ids=("id", list)
            )
            .reset_index()
        )
        grouped.rename(
            columns={
                "community_x": "community",
                "level_x": "level",
                "parent_x": "parent",
            },
            inplace=True,
        )
        all_grouped = pd.concat([
            all_grouped,
            grouped.loc[
                :, ["community", "level", "parent", "relationship_ids"]
            ],
        ])

    # deduplicate the relationship_ids lists
    all_grouped["relationship_ids"] = all_grouped["relationship_ids"].apply(
        lambda x: sorted(set(x)) if x else []
    )
    
    # 通过entity_ids和relationship_ids推断text_unit_ids
    # 合并entity_ids以获取每个社区的实体
    all_grouped_with_entities = all_grouped.merge(entity_ids, on="community", how="inner")
    
    # 为每个社区收集text_unit_ids
    def collect_text_unit_ids(row):
        """从社区的实体和关系中收集所有text_unit_ids"""
        text_units = set()
        
        # 辅助函数：安全地添加text_unit_ids到集合
        def add_text_units_safely(text_unit_data):
            """安全地将text_unit_ids添加到集合中"""
            if text_unit_data is None:
                return
            
            # 处理numpy数组
            if hasattr(text_unit_data, '__iter__') and not isinstance(text_unit_data, str):
                # 是可迭代对象（列表、数组等），但不是字符串
                try:
                    for item in text_unit_data:
                        if item is not None:
                            # 确保item是可哈希的（字符串或数字）
                            text_units.add(str(item) if not isinstance(item, str) else item)
                except (TypeError, ValueError):
                    # 如果迭代失败，尝试作为单个值处理
                    text_units.add(str(text_unit_data))
            else:
                # 单个值
                text_units.add(str(text_unit_data) if not isinstance(text_unit_data, str) else text_unit_data)
        
        # 1. 从实体的text_unit_ids收集
        entity_ids_list = row.get("entity_ids", [])
        if entity_ids_list:
            for entity_id in entity_ids_list:
                entity_rows = entities[entities["id"] == entity_id]
                if not entity_rows.empty:
                    entity_text_units = entity_rows.iloc[0].get("text_unit_ids", [])
                    add_text_units_safely(entity_text_units)
        
        # 2. 从关系的text_unit_ids收集
        relationship_ids_list = row.get("relationship_ids", [])
        if relationship_ids_list:
            for rel_id in relationship_ids_list:
                rel_rows = relationships[relationships["id"] == rel_id]
                if not rel_rows.empty:
                    rel_text_units = rel_rows.iloc[0].get("text_unit_ids", [])
                    add_text_units_safely(rel_text_units)
        
        # 返回排序后的列表
        return sorted(list(text_units))
    
    logger.info("Inferring text_unit_ids from entity_ids and relationship_ids for each community...")
    all_grouped_with_entities["text_unit_ids"] = all_grouped_with_entities.apply(
        collect_text_unit_ids, axis=1
    )
    
    # 移除临时添加的entity_ids列（稍后会重新合并）
    all_grouped = all_grouped_with_entities.drop(columns=["entity_ids"])

    # join it all up and add some new fields
    final_communities = all_grouped.merge(entity_ids, on="community", how="inner")
    final_communities["id"] = [str(uuid4()) for _ in range(len(final_communities))]
    final_communities["human_readable_id"] = final_communities["community"]
    final_communities["title"] = "Community " + final_communities["community"].astype(
        str
    )
    final_communities["parent"] = final_communities["parent"].astype(int)
    # collect the children so we have a tree going both ways
    parent_grouped = cast(
        "pd.DataFrame",
        final_communities.groupby("parent").agg(children=("community", "unique")),
    )
    final_communities = final_communities.merge(
        parent_grouped,
        left_on="community",
        right_on="parent",
        how="left",
    )
    # replace NaN children with empty list
    final_communities["children"] = final_communities["children"].apply(
        lambda x: x if isinstance(x, np.ndarray) else []  # type: ignore
    )
    # add fields for incremental update tracking
    final_communities["period"] = datetime.now(timezone.utc).date().isoformat()
    final_communities["size"] = final_communities.loc[:, "entity_ids"].apply(len)

    return final_communities.loc[
        :,
        COMMUNITIES_FINAL_COLUMNS,
    ]
