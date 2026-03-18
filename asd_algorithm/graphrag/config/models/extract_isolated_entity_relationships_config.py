# Copyright (c) 2024 Microsoft Corporation.
# Licensed under the MIT License

"""Parameterization settings for isolated entity relationship extraction configuration."""

from pathlib import Path

from pydantic import BaseModel, Field

from graphrag.config.defaults import graphrag_config_defaults
from graphrag.config.models.language_model_config import LanguageModelConfig


class ExtractIsolatedEntityRelationshipsConfig(BaseModel):
    """Configuration section for isolated entity relationship extraction."""

    model_id: str = Field(
        description="The model ID to use for isolated entity relationship extraction.",
        default="default_chat_model",
    )
    prompt: str | None = Field(
        description="The isolated entity relationship extraction prompt to use.",
        default="prompts/extract_isolated_entity_relationships.txt",
    )
    max_gleanings: int = Field(
        description="The maximum number of relationship gleanings to use.",
        default=1,
    )
    tuple_delimiter: str = Field(
        description="The tuple delimiter to use in the prompt.",
        default="<|>",
    )
    record_delimiter: str = Field(
        description="The record delimiter to use in the prompt.",
        default="##",
    )
    completion_delimiter: str = Field(
        description="The completion delimiter to use in the prompt.",
        default="<|COMPLETE|>",
    )
    strategy: dict | None = Field(
        description="Override the default isolated entity relationship extraction strategy",
        default=None,
    )

    def resolved_strategy(
        self, root_dir: str, model_config: LanguageModelConfig
    ) -> dict:
        """Get the resolved isolated entity relationship extraction strategy."""
        return self.strategy or {
            "llm": model_config.model_dump(),
            "extraction_prompt": (Path(root_dir) / self.prompt).read_text(
                encoding="utf-8"
            )
            if self.prompt
            else None,
            "max_gleanings": self.max_gleanings,
            "tuple_delimiter": self.tuple_delimiter,
            "record_delimiter": self.record_delimiter,
            "completion_delimiter": self.completion_delimiter,
        }
