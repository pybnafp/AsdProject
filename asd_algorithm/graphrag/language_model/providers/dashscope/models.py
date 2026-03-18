"""
DashScope Language Model provider for GraphRAG.

This module provides DashScope (Alibaba Cloud) language model integration
for GraphRAG query operations.
"""

import asyncio
import logging
from typing import Any, AsyncGenerator, Generator

from dashscope import Generation

from graphrag.language_model.response.base import (
    BaseModelOutput,
    BaseModelResponse,
    ModelResponse,
)
from graphrag.language_model.protocol.base import ChatModel

logger = logging.getLogger(__name__)


class DashScopeChatModel:
    """
    DashScope Chat Model provider for GraphRAG.
    
    This class implements the ChatModel protocol using DashScope's Generation API.
    """

    def __init__(
        self,
        *,
        name: str,
        config: Any = None,
        api_key: str | None = None,
        model: str = "qwen-plus",
        max_tokens: int = 1024,
        system_message: str | None = None,
        callbacks: Any | None = None,
        cache: Any | None = None,
        **kwargs: Any,
    ) -> None:
        """
        Initialize DashScope Chat Model.
        
        Args:
            name: Model name identifier
            config: Language model configuration
            api_key: DashScope API key
            model: Model name (default: qwen-plus)
            max_tokens: Maximum tokens for response
            system_message: System message for the model
            callbacks: Optional callbacks object (ignored by DashScope provider)
            cache: Optional cache object (ignored by DashScope provider)
            **kwargs: Additional keyword args for compatibility (ignored)
        """
        self.name = name
        self.config = config
        # Resolve api_key: prioritize explicit param, then config.api_key
        self.api_key = api_key if api_key else (
            getattr(config, "api_key", None) if config is not None else None
        )
        self.model = model
        self.max_tokens = max_tokens
        self.system_message = system_message
        # Store for potential future use; currently not used
        self._callbacks = callbacks
        self._cache = cache

    async def achat(
        self, prompt: str, history: list | None = None, **kwargs: Any
    ) -> ModelResponse:
        """
        Chat with the DashScope model using the given prompt.

        Args:
            prompt: The prompt to chat with.
            history: Conversation history (包含系统提示词)
            **kwargs: Additional arguments including json, json_model, name, etc.

        Returns:
            The response from the model.
        """
        try:
            # 提取JSON相关参数
            json_mode = kwargs.get('json', False)
            json_model = kwargs.get('json_model', None)
            name = kwargs.get('name', 'dashscope_chat')
            
            # 组装提示词
            messages = []

            # 常量：安全的系统消息默认值
            default_system_message = "你是一个专业的AI助手，请根据提供的信息进行客观、简洁的回答。"

            # 如果有历史对话，添加到消息中（过滤无效项）
            if history:
                for msg in history:
                    if (
                        isinstance(msg, dict)
                        and msg.get("role") in ("system", "user", "assistant")
                        and ("content" in msg)
                        and (msg.get("content") is not None)
                    ):
                        messages.append({
                            "role": msg["role"],
                            "content": str(msg["content"]),
                        })
            else:
                # 如果没有历史对话，使用默认系统消息（避免 None）
                messages.append({
                    "role": "system",
                    "content": self.system_message or default_system_message,
                })

            # 添加用户消息（避免 None）
            messages.append({
                "role": "user",
                "content": prompt if (prompt is not None and str(prompt).strip() != "") else "请回答。",
            })
            
            # 调用 DashScope API
            # 当json_mode=True时，使用response_format确保输出格式化的JSON
            call_kwargs = {
                "api_key": self.api_key,
                "model": self.model,
                "messages": messages,
                "result_format": "message",
                "max_tokens": self.max_tokens
            }
            
            # 如果启用JSON模式，添加response_format参数以确保输出格式化的JSON
            if json_mode:
                call_kwargs["response_format"] = {"type": "json_object"}
            
            response = Generation.call(**call_kwargs)
            
            # 验证响应对象
            if response is None:
                logger.error("DashScope API returned None response")
                return BaseModelResponse(
                    output=BaseModelOutput(
                        content="抱歉，API 返回了空响应。",
                        full_response={"error": "API returned None response"},
                    ),
                    parsed_response=None,
                    history=messages,
                    cache_hit=False,
                    tool_calls=[],
                )
            
            # 提取回答
            if hasattr(response, 'status_code') and response.status_code == 200:
                if hasattr(response, 'output') and response.output and hasattr(response.output, 'choices'):
                    answer = response.output.choices[0].message.content
                    
                    # 处理JSON模式响应
                    parsed_response = None
                    if json_mode and json_model and answer:
                        try:
                            # 尝试解析JSON响应
                            import json
                            parsed_json = json.loads(answer)
                            # 使用json_model验证和转换
                            if hasattr(json_model, 'model_validate'):
                                parsed_response = json_model.model_validate(parsed_json)
                            elif hasattr(json_model, 'parse_obj'):
                                parsed_response = json_model.parse_obj(parsed_json)
                            else:
                                # 如果json_model没有验证方法，直接使用解析的JSON
                                parsed_response = parsed_json
                        except (json.JSONDecodeError, ValueError, TypeError) as e:
                            logger.warning(f"Failed to parse JSON response for {name}: {e}")
                            logger.debug(f"Raw response content: {answer}")
                            
                            # 尝试修复JSON格式问题
                            try:
                                import json
                                import re
                                
                                # 修复常见的JSON格式问题
                                fixed_answer = answer
                                
                                # 1. 处理未闭合的字符串：找到最后一个完整的JSON对象
                                # 尝试找到最后一个完整的 findings 数组项
                                if '"findings"' in fixed_answer:
                                    # 找到 findings 数组的开始
                                    findings_start = fixed_answer.find('"findings"')
                                    if findings_start != -1:
                                        # 尝试找到最后一个完整的 finding 对象
                                        # 查找所有 "}," 或 "}" 在 findings 之后的位置
                                        findings_section = fixed_answer[findings_start:]
                                        
                                        # 尝试多个截断点
                                        for i in range(len(findings_section) - 1, -1, -1):
                                            if findings_section[i:i+2] == '},':
                                                # 尝试在这里截断并补全JSON
                                                test_json = fixed_answer[:findings_start + i + 1] + ']}'
                                                try:
                                                    test_parsed = json.loads(test_json)
                                                    # 如果解析成功，使用这个修复后的JSON
                                                    fixed_answer = test_json
                                                    logger.info(f"Successfully repaired JSON by truncating at position {findings_start + i + 1}")
                                                    break
                                                except:
                                                    continue
                                
                                # 2. 尝试解析修复后的JSON
                                parsed_json = json.loads(fixed_answer)
                                
                                # 使用json_model验证和转换
                                if hasattr(json_model, 'model_validate'):
                                    parsed_response = json_model.model_validate(parsed_json)
                                elif hasattr(json_model, 'parse_obj'):
                                    parsed_response = json_model.parse_obj(parsed_json)
                                else:
                                    parsed_response = parsed_json
                                
                                logger.info(f"Successfully parsed repaired JSON for {name}")
                            except Exception as repair_e:
                                logger.debug(f"Failed to repair JSON: {repair_e}")
                                parsed_response = None
                    
                    return BaseModelResponse(
                        output=BaseModelOutput(
                            content=answer,
                            full_response=None,
                        ),
                        parsed_response=parsed_response,
                        history=messages,
                        cache_hit=False,
                        tool_calls=[],
                    )
                else:
                    logger.error("DashScope API response missing output or choices")
                    return BaseModelResponse(
                        output=BaseModelOutput(
                            content="抱歉，API 响应格式不正确。",
                            full_response=None,
                        ),
                        parsed_response=None,
                        history=messages,
                        cache_hit=False,
                        tool_calls=[],
                    )
            else:
                error_message = getattr(response, 'message', '未知错误')
                logger.error(f"DashScope API error: {error_message}")
                return BaseModelResponse(
                    output=BaseModelOutput(
                        content="抱歉，我无法处理您的请求。",
                        full_response=None,
                    ),
                    parsed_response=None,
                    history=messages,
                    cache_hit=False,
                    tool_calls=[],
                )
                
        except Exception as e:
            logger.error(f"Error in DashScope achat: {e}")
            return BaseModelResponse(
                output=BaseModelOutput(
                    content="抱歉，处理您的请求时出现了错误。",
                    full_response={"error": str(e)},
                ),
                parsed_response=None,
                history=[],
                cache_hit=False,
                tool_calls=[],
            )

    async def achat_stream(
        self, prompt: str, history: list | None = None, **kwargs: Any
    ) -> AsyncGenerator[str, None]:
        """
        Stream chat with the DashScope model using the given prompt.

        Args:
            prompt: The prompt to chat with.
            history: Conversation history (包含系统提示词)
            **kwargs: Additional arguments (ignored for now)

        Yields:
            String chunks of the response.
        """
        try:
            # 组装提示词
            messages = []

            default_system_message = "你是一个专业的AI助手，请根据提供的信息进行客观、简洁的回答。"

            if history:
                for msg in history:
                    if (
                        isinstance(msg, dict)
                        and msg.get("role") in ("system", "user", "assistant")
                        and ("content" in msg)
                        and (msg.get("content") is not None)
                    ):
                        messages.append({
                            "role": msg["role"],
                            "content": str(msg["content"]),
                        })
            else:
                messages.append({
                    "role": "system",
                    "content": self.system_message or default_system_message,
                })

            messages.append({
                "role": "user",
                "content": prompt if (prompt is not None and str(prompt).strip() != "") else "请回答。",
            })
            
            # 调用 DashScope API (流式)
            response = Generation.call(
                api_key=self.api_key,
                model=self.model,
                messages=messages,
                result_format="message",
                max_tokens=self.max_tokens,
                stream=True
            )
            
            # 验证响应对象
            if response is None:
                logger.error("DashScope API returned None response for stream")
                yield "抱歉，API 返回了空响应。"
                return
            
            # 流式返回结果 - 当 stream=True 时，response 是一个生成器
            try:
                accumulated_content = ""
                for chunk in response:
                    if chunk is None:
                        continue
                    
                    # 检查 chunk 是否有错误
                    if hasattr(chunk, 'status_code') and chunk.status_code != 200:
                        error_message = getattr(chunk, 'message', '未知错误')
                        logger.error(f"DashScope API error in chunk: {error_message}")
                        yield "抱歉，我无法处理您的请求。"
                        return
                    
                    # 提取内容
                    chunk_content = ""
                    if hasattr(chunk, 'output') and chunk.output and hasattr(chunk.output, 'choices'):
                        if chunk.output.choices and len(chunk.output.choices) > 0:
                            chunk_content = chunk.output.choices[0].message.content or ""
                    elif hasattr(chunk, 'output') and chunk.output:
                        # 尝试直接获取内容
                        if hasattr(chunk.output, 'text'):
                            chunk_content = chunk.output.text or ""
                        elif hasattr(chunk.output, 'content'):
                            chunk_content = chunk.output.content or ""
                    
                    if chunk_content:
                        # 检查是否是增量内容还是完整内容
                        if chunk_content.startswith(accumulated_content):
                            # 这是增量内容，只输出新增部分
                            new_content = chunk_content[len(accumulated_content):]
                            if new_content:
                                yield new_content
                                accumulated_content = chunk_content
                        else:
                            # 这可能是完整内容，检查是否与之前的内容不同
                            if chunk_content != accumulated_content:
                                # 如果内容完全不同，输出整个内容
                                yield chunk_content
                                accumulated_content = chunk_content
                            # 如果内容相同，跳过（避免重复输出）
                            
            except Exception as stream_error:
                logger.error(f"Error processing stream chunks: {stream_error}")
                yield "抱歉，处理流式响应时出现了错误。"
                
        except Exception as e:
            logger.error(f"Error in DashScope achat_stream: {e}")
            yield "抱歉，处理您的请求时出现了错误。"

    def chat(self, prompt: str, history: list | None = None, **kwargs: Any) -> ModelResponse:
        """
        Synchronous chat with the DashScope model.

        Args:
            prompt: The prompt to chat with.
            history: Conversation history
            **kwargs: Additional arguments including json, json_model, name, etc.

        Returns:
            The response from the model.
        """
        return asyncio.run(self.achat(prompt, history=history, **kwargs))

    def chat_stream(
        self, prompt: str, history: list | None = None, **kwargs: Any
    ) -> Generator[str, None, None]:
        """
        Synchronous stream chat with the DashScope model.

        Args:
            prompt: The prompt to chat with.
            history: Conversation history
            **kwargs: Additional arguments

        Yields:
            String chunks of the response.
        """
        async def _async_generator():
            async for chunk in self.achat_stream(prompt, history=history, **kwargs):
                yield chunk
        
        # 运行异步生成器
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            async_gen = _async_generator()
            while True:
                try:
                    chunk = loop.run_until_complete(async_gen.__anext__())
                    yield chunk
                except StopAsyncIteration:
                    break
        finally:
            loop.close()
