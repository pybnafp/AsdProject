"""
Indexing API for GraphRAG.
"""

import asyncio
import sys
import logging
from typing import Any
from pathlib import Path
import json

from graphrag.config.load_config import load_config
from graphrag.index.validate_config import validate_config_names
from graphrag.callbacks.noop_workflow_callbacks import NoopWorkflowCallbacks
from graphrag.callbacks.workflow_callbacks import WorkflowCallbacks
from graphrag.config.enums import CacheType, IndexingMethod
from graphrag.config.models.graph_rag_config import GraphRagConfig
from graphrag.index.run.run_pipeline import run_pipeline
from graphrag.index.run.utils import create_callback_chain
from graphrag.index.typing.pipeline_run_result import PipelineRunResult
from graphrag.index.typing.workflow import WorkflowFunction
from graphrag.index.workflows.factory import PipelineFactory
from graphrag.logger.standard_logging import init_loggers

logger = logging.getLogger(__name__)


async def build_index(
    config: GraphRagConfig,
    method: IndexingMethod | str = IndexingMethod.Standard,
    is_update_run: bool = False,
    memory_profile: bool = False,
    callbacks: list[WorkflowCallbacks] | None = None,
    additional_context: dict[str, Any] | None = None,
    verbose: bool = False,
) -> list[PipelineRunResult]:
    """Run the pipeline with the given configuration.

    Parameters
    ----------
    config : GraphRagConfig
        The configuration.
    method : IndexingMethod default=IndexingMethod.Standard
        Styling of indexing to perform (full LLM, NLP + LLM, etc.).
    memory_profile : bool
        Whether to enable memory profiling.
    callbacks : list[WorkflowCallbacks] | None default=None
        A list of callbacks to register.
    additional_context : dict[str, Any] | None default=None
        Additional context to pass to the pipeline run. This can be accessed in the pipeline state under the 'additional_context' key.

    Returns
    -------
    list[PipelineRunResult]
        The list of pipeline run results
    """
    init_loggers(config=config, verbose=verbose)

    # Create callbacks for pipeline lifecycle events if provided
    workflow_callbacks = (
        create_callback_chain(callbacks) if callbacks else NoopWorkflowCallbacks()
    )

    outputs: list[PipelineRunResult] = []

    if memory_profile:
        logger.warning("New pipeline does not yet support memory profiling.")

    logger.info("Initializing indexing pipeline...")
    # todo: this could propagate out to the cli for better clarity, but will be a breaking api change
    method = _get_method(method, is_update_run)
    pipeline = PipelineFactory.create_pipeline(config, method)

    workflow_callbacks.pipeline_start(pipeline.names())

    async for output in run_pipeline(
        pipeline,
        config,
        callbacks=workflow_callbacks,
        is_update_run=is_update_run,
        additional_context=additional_context,
    ):
        outputs.append(output)
        if output.errors and len(output.errors) > 0:
            logger.error("Workflow %s completed with errors", output.workflow)
        else:
            logger.info("Workflow %s completed successfully", output.workflow)
        logger.debug(str(output.result))

    workflow_callbacks.pipeline_end(outputs)
    return outputs


def update_index(
    root_dir: Path,
    method: IndexingMethod,
    verbose: bool,
    memprofile: bool,
    cache: bool,
    config_filepath: Path | None,
    skip_validation: bool,
    output_dir: Path | None,
    dry_run: bool = False,
):
    """Run the pipeline with the given config."""
    # cli_overrides = {}
    # if output_dir:
    #     cli_overrides["output.base_dir"] = str(output_dir)
    #     cli_overrides["reporting.base_dir"] = str(output_dir)
    #     cli_overrides["update_index_output.base_dir"] = str(output_dir)

    config = load_config(root_dir, config_filepath)


    # Initialize loggers and reporting config
    # Ensure logs are written to a stable file under reporting.base_dir (e.g., logs/update_index.log)
    init_loggers(
        config=config,
        verbose=verbose,
        filename="update_index.log",
    )

    if not cache:
        config.cache.type = CacheType.none

    if not skip_validation:
        validate_config_names(config)

    logger.info("Starting pipeline run. %s", dry_run)
    logger.info(
        "Using default configuration: %s",
        redact(config.model_dump()),
    )

    if dry_run:
        logger.info("Dry run complete, exiting...", True)
        sys.exit(0)

    _register_signal_handlers()

    outputs = asyncio.run(
        build_index(
            config=config,
            method=method,
            is_update_run=True, 
            memory_profile=memprofile,
            callbacks=None,
            verbose=verbose,
        )
    )
    encountered_errors = any(
        output.errors and len(output.errors) > 0 for output in outputs
    )

    if encountered_errors:
        logger.error(
            "Errors occurred during the pipeline run, see logs for more details."
        )
    else:
        logger.info("All workflows completed successfully.")

    sys.exit(1 if encountered_errors else 0)


def register_workflow_function(name: str, workflow: WorkflowFunction):
    """Register a custom workflow function. You can then include the name in the settings.yaml workflows list."""
    PipelineFactory.register(name, workflow)


def _get_method(method: IndexingMethod | str, is_update_run: bool) -> str:
    m = method.value if isinstance(method, IndexingMethod) else method
    # return f"{m}-update" if is_update_run else m
    return m


def _register_signal_handlers():
    import signal

    def handle_signal(signum, _):
        # Handle the signal here
        logger.debug(f"Received signal {signum}, exiting...")  # noqa: G004
        for task in asyncio.all_tasks():
            task.cancel()
        logger.debug("All tasks cancelled. Exiting...")

    # Register signal handlers for SIGINT and SIGHUP
    signal.signal(signal.SIGINT, handle_signal)

    if sys.platform != "win32":
        signal.signal(signal.SIGHUP, handle_signal)


def redact(config: dict) -> str:
    """Sanitize secrets in a config object."""

    # Redact any sensitive configuration
    def redact_dict(config: dict) -> dict:
        if not isinstance(config, dict):
            return config

        result = {}
        for key, value in config.items():
            if key in {
                "api_key",
                "connection_string",
                "container_name",
                "organization",
            }:
                if value is not None:
                    result[key] = "==== REDACTED ===="
            elif isinstance(value, dict):
                result[key] = redact_dict(value)
            elif isinstance(value, list):
                result[key] = [redact_dict(i) for i in value]
            else:
                result[key] = value
        return result

    redacted_dict = redact_dict(config)
    return json.dumps(redacted_dict, indent=4)
