"""A module containing run_workflow method definition."""

import logging
from typing import Any

import pandas as pd

from graphrag.config.models.embed_graph_config import EmbedGraphConfig
from graphrag.config.models.graph_rag_config import GraphRagConfig
from graphrag.index.operations.create_graph import create_graph
from graphrag.index.operations.finalize_entities import finalize_entities
from graphrag.index.operations.finalize_relationships import finalize_relationships
from graphrag.index.operations.snapshot_graphml import snapshot_graphml
from graphrag.index.typing.context import PipelineRunContext
from graphrag.index.typing.workflow import WorkflowFunctionOutput
from graphrag.utils.storage import load_table_from_storage, write_table_to_storage

logger = logging.getLogger(__name__)


async def run_workflow(
    config: GraphRagConfig,
    context: PipelineRunContext,
) -> WorkflowFunctionOutput:
    """All the steps to create the base entity graph."""
    logger.info("Workflow started: finalize_graph")
    entities = await load_table_from_storage("entities", context.output_storage)
    relationships = await load_table_from_storage(
        "relationships", context.output_storage
    )
    
    # 加载文本单元用于孤立实体关系提取
    text_units = None
    try:
        text_units = await load_table_from_storage("text_units", context.output_storage)
    except Exception as e:
        logger.warning(f"Could not load text_units for isolated entity relationship extraction: {e}")
    
    # 获取孤立实体关系提取配置
    isolated_entity_config = None
    try:
        isolated_entity_config = config.extract_isolated_entity_relationships.resolved_strategy(
            config.root_dir, 
            config.get_language_model_config(config.extract_isolated_entity_relationships.model_id)
        )
    except Exception as e:
        logger.warning(f"Could not load isolated entity relationship extraction config: {e}")

    final_entities, final_relationships = await finalize_graph(
        entities,
        relationships,
        embed_config=config.embed_graph,
        layout_enabled=config.umap.enabled,
        config=isolated_entity_config,
        cache=context.cache,
        text_units=text_units,
    )

    await write_table_to_storage(final_entities, "entities", context.output_storage)
    await write_table_to_storage(
        final_relationships, "relationships", context.output_storage
    )

    if config.snapshots.graphml:
        # todo: extract graphs at each level, and add in meta like descriptions
        graph = create_graph(final_relationships, edge_attr=["weight"])

        await snapshot_graphml(
            graph,
            name="graph",
            storage=context.output_storage,
        )

    logger.info("Workflow completed: finalize_graph")
    return WorkflowFunctionOutput(
        result={
            "entities": entities,
            "relationships": relationships,
        }
    )


async def finalize_graph(
    entities: pd.DataFrame,
    relationships: pd.DataFrame,
    embed_config: EmbedGraphConfig | None = None,
    layout_enabled: bool = False,
    config: Any = None,
    cache: Any = None,
    text_units: pd.DataFrame | None = None,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """All the steps to finalize the entity and relationship formats."""
    final_entities = await finalize_entities(
        entities, relationships, embed_config, layout_enabled, config=config, cache=cache, text_units=text_units
    )
    # 传递实体数据确保所有实体都有对应的关系
    final_relationships = await finalize_relationships(relationships, entities, config=config, cache=cache, text_units=text_units)
    return (final_entities, final_relationships)
