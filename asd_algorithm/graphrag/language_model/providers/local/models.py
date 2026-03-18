"""
Local Language Model provider for GraphRAG.

This module provides local Qwen3-4B model integration for GraphRAG operations.
"""

import asyncio
import gc
import logging
import os
import re
from typing import Any, AsyncGenerator, Generator, List, Optional

import torch
from transformers import AutoModelForCausalLM, AutoTokenizer

from graphrag.language_model.response.base import (
    BaseModelOutput,
    BaseModelResponse,
    ModelResponse,
)
from graphrag.language_model.protocol.base import ChatModel

logger = logging.getLogger(__name__)


def clear_cuda_cache():
    """清理CUDA缓存"""
    if torch.cuda.is_available():
        torch.cuda.empty_cache()
        gc.collect()


def build_messages(user_prompt: str, system_prompt: Optional[str] = None, history: Optional[List[List[str]]] = None):
    messages = []
    if system_prompt:
        messages.append({"role": "system", "content": system_prompt})
    if history:
        for user_turn, assistant_turn in history:
            messages.append({"role": "user", "content": user_turn})
            messages.append({"role": "assistant", "content": assistant_turn})
    messages.append({"role": "user", "content": user_prompt})
    return messages


class LocalQwen3ChatModel:
    """
    Local Qwen3-4B Chat Model provider for GraphRAG.
    
    This class implements the ChatModel protocol using local Qwen3-4B model.
    """

    def __init__(
        self,
        *,
        name: str,
        config: Any = None,
        model_dir: str | None = None,
        dtype: str | None = None,
        max_tokens: int = 2048,
        system_message: str | None = None,
        callbacks: Any | None = None,
        cache: Any | None = None,
        **kwargs: Any,
    ) -> None:
        """
        Initialize Local Qwen3 Chat Model.
        
        Args:
            name: Model name identifier
            config: Language model configuration
            model_dir: Local model directory path
            dtype: Model precision (fp16/bf16/fp32)
            max_tokens: Maximum tokens for response
            system_message: System message for the model
            callbacks: Optional callbacks object (ignored by local provider)
            cache: Optional cache object (ignored by local provider)
            **kwargs: Additional keyword args for compatibility (ignored)
        """
        self.name = name
        self.config = config
        
        # 默认模型目录
        if model_dir is None:
            # 从配置中获取或使用默认路径
            model_dir = getattr(config, "model_dir", None) if config else None
            if model_dir is None:
                # 使用相对于项目根目录的默认路径
                current_dir = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
                model_dir = os.path.join(current_dir, "model", "Qwen3-4B")
        
        self.model_dir = model_dir
        self.dtype = dtype
        self.max_tokens = max_tokens
        self.system_message = system_message or "你是一个专业的AI助手，请根据提供的信息进行客观、简洁的回答。"
        
        # Store for potential future use; currently not used
        self._callbacks = callbacks
        self._cache = cache
        
        # 延迟加载模型
        self._model = None
        self._tokenizer = None
        self._device = None

    def _load_model(self):
        """延迟加载模型和tokenizer"""
        if self._model is not None and self._tokenizer is not None:
            return
            
        if not os.path.isdir(self.model_dir):
            raise FileNotFoundError(f"模型目录不存在: {self.model_dir}")

        self._device = "cuda" if torch.cuda.is_available() else "cpu"
        logger.info(f"检测到设备: {self._device}")
        
        if self._device == "cuda":
            logger.info(f"CUDA 设备数量: {torch.cuda.device_count()}")
            logger.info(f"当前 CUDA 设备: {torch.cuda.current_device()}")
            logger.info(f"GPU 名称: {torch.cuda.get_device_name()}")
            logger.info(f"GPU 显存: {torch.cuda.get_device_properties(0).total_memory / 1024**3:.1f} GB")

        torch_dtype = None
        if self.dtype:
            dtype_lower = self.dtype.lower()
            if dtype_lower in ("fp16", "float16", "half"):
                torch_dtype = torch.float16
            elif dtype_lower in ("bf16", "bfloat16"):
                torch_dtype = torch.bfloat16
            elif dtype_lower in ("fp32", "float32"):
                torch_dtype = torch.float32
        else:
            # CUDA 设备默认使用 fp16 以节省显存
            if self._device == "cuda":
                torch_dtype = torch.float16
                logger.info("CUDA 设备默认使用 fp16 精度")

        logger.info("正在加载 tokenizer...")
        self._tokenizer = AutoTokenizer.from_pretrained(self.model_dir, use_fast=True)
        
        # 设置 pad_token
        if self._tokenizer.pad_token is None:
            self._tokenizer.pad_token = self._tokenizer.eos_token
            logger.info("设置 pad_token 为 eos_token")

        logger.info("正在加载模型...")
        model_kwargs = {}
        
        if self._device == "cuda":
            # CUDA 优化设置
            model_kwargs["torch_dtype"] = torch_dtype
            
            # 检查是否安装了 accelerate
            try:
                import accelerate  # type: ignore
                model_kwargs["device_map"] = "auto"
                logger.info("使用 accelerate 进行自动设备映射")
            except ImportError:
                logger.info("未安装 accelerate，使用手动设备映射")
                # 不使用 device_map，稍后手动移动到 GPU
            
            # 启用内存优化
            try:
                torch.backends.cuda.matmul.allow_tf32 = True
                torch.backends.cudnn.allow_tf32 = True
                logger.info("启用 TF32 优化")
            except Exception:
                pass
        else:
            # CPU 设置
            if torch_dtype is None:
                torch_dtype = torch.float32
            model_kwargs["torch_dtype"] = torch_dtype

        try:
            self._model = AutoModelForCausalLM.from_pretrained(self.model_dir, **model_kwargs)
        except Exception as e:
            logger.error(f"模型加载失败: {e}")
            # 尝试使用更兼容的参数
            fallback_kwargs = {"torch_dtype": torch_dtype}
            if self._device == "cuda":
                try:
                    import accelerate  # type: ignore
                    fallback_kwargs["device_map"] = "auto"
                except ImportError:
                    pass
            logger.info("尝试使用兼容参数重新加载模型...")
            self._model = AutoModelForCausalLM.from_pretrained(self.model_dir, **fallback_kwargs)

        # 手动设备映射（如果没有使用 accelerate）
        if self._device == "cuda" and "device_map" not in model_kwargs:
            try:
                self._model = self._model.cuda()
                logger.info("手动将模型移动到 GPU")
            except Exception as e:
                logger.error(f"移动到 GPU 失败: {e}")
                logger.info("回退到 CPU 模式")
                self._device = "cpu"
                self._model = self._model.to(torch.float32)
        elif self._device == "cpu":
            self._model = self._model.to(torch.float32)
        
        # 设置为评估模式
        self._model.eval()
        
        if self._device == "cuda":
            logger.info(f"模型已加载到 GPU，显存使用: {torch.cuda.memory_allocated() / 1024**3:.2f} GB")

    def _generate_answer(
        self,
        prompt: str,
        history: Optional[List[List[str]]] = None,
        max_new_tokens: int = None,
        temperature: float = 0.7,
        top_p: float = 0.9,
        do_sample: bool = True,
        enable_think: bool = False,
    ) -> str:
        """生成回答的内部方法"""
        if max_new_tokens is None:
            max_new_tokens = self.max_tokens
            
        messages = build_messages(prompt, system_prompt=self.system_message, history=history)

        try:
            # 优先使用 chat 模板（适配 Qwen3 的对话格式）
            inputs = self._tokenizer.apply_chat_template(
                messages,
                add_generation_prompt=True,
                return_tensors="pt",
            )
            # 某些 tokenizer 会直接返回 Tensor 而不是 BatchEncoding
            if isinstance(inputs, torch.Tensor):
                inputs = {"input_ids": inputs}
        except Exception as e:
            logger.warning(f"Chat template 失败，使用回退方案: {e}")
            # 回退到直接拼接文本
            joined = []
            for m in messages:
                role = m.get("role", "user")
                content = m.get("content", "")
                joined.append(f"{role}: {content}")
            joined.append("assistant:")
            inputs = self._tokenizer("\n".join(joined), return_tensors="pt")

        # 兼容 inputs 可能是 Tensor 的情况
        if isinstance(inputs, torch.Tensor):
            inputs = {"input_ids": inputs}
        
        # 将输入移动到模型设备
        device = next(self._model.parameters()).device
        inputs = {k: v.to(device) for k, v in inputs.items()}
        
        # 检查输入长度
        input_length = inputs["input_ids"].shape[1]
        if input_length > 4000:  # 如果输入过长，给出警告
            logger.warning(f"输入长度 {input_length} 较长，可能影响生成质量")

        generation_config = dict(
            max_new_tokens=max_new_tokens,
            temperature=temperature,
            top_p=top_p,
            do_sample=do_sample,
            eos_token_id=getattr(self._tokenizer, "eos_token_id", None),
            pad_token_id=getattr(self._tokenizer, "pad_token_id", None) or getattr(self._tokenizer, "eos_token_id", None),
            repetition_penalty=1.1,  # 减少重复
            no_repeat_ngram_size=3,  # 避免3-gram重复
        )

        # CUDA 优化：使用 torch.cuda.amp 进行混合精度推理
        if device.type == "cuda":
            try:
                with torch.cuda.amp.autocast():
                    with torch.no_grad():
                        outputs = self._model.generate(**inputs, **generation_config)
            except Exception:
                # 如果混合精度失败，回退到普通推理
                with torch.no_grad():
                    outputs = self._model.generate(**inputs, **generation_config)
        else:
            with torch.no_grad():
                outputs = self._model.generate(**inputs, **generation_config)

        # 截取新生成的部分，避免回显提示词
        if outputs.shape[0] == 1 and inputs.get("input_ids") is not None:
            new_tokens = outputs[0][input_length:]
            text = self._tokenizer.decode(new_tokens, skip_special_tokens=True)
        else:
            text = self._tokenizer.decode(outputs[0], skip_special_tokens=True)

        text = text.strip()

        # 非 think 模式：移除思考过程标记片段
        if not enable_think and text:
            patterns = [
                r"<think>[\s\S]*?</think>",
                r"<\|think\|>[\s\S]*?</think>",
                r"<\|im_think\|>[\s\S]*?<\|im_end\|>",
                r"<THINK>[\s\S]*?</THINK>",
            ]
            for pat in patterns:
                text = re.sub(pat, "", text, flags=re.IGNORECASE)
            text = text.strip()

        return text

    async def achat(
        self, prompt: str, history: list | None = None, **kwargs: Any
    ) -> ModelResponse:
        """
        Chat with the local Qwen3 model using the given prompt.

        Args:
            prompt: The prompt to chat with.
            history: Conversation history
            **kwargs: Additional arguments including json, json_model, name, etc.

        Returns:
            The response from the model.
        """
        try:
            # 延迟加载模型
            self._load_model()
            
            # 提取参数
            json_mode = kwargs.get('json', False)
            json_model = kwargs.get('json_model', None)
            name = kwargs.get('name', 'local_qwen3_chat')
            max_new_tokens = kwargs.get('max_new_tokens', self.max_tokens)
            temperature = kwargs.get('temperature', 0.7)
            top_p = kwargs.get('top_p', 0.9)
            do_sample = kwargs.get('do_sample', True)
            enable_think = kwargs.get('enable_think', False)
            
            # 生成回答
            answer = self._generate_answer(
                prompt=prompt,
                history=history,
                max_new_tokens=max_new_tokens,
                temperature=temperature,
                top_p=top_p,
                do_sample=do_sample,
                enable_think=enable_think,
            )
            
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
                    parsed_response = None
            
            # 构建历史消息
            messages = []
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
                    "content": self.system_message,
                })
            
            messages.append({
                "role": "user",
                "content": prompt if (prompt is not None and str(prompt).strip() != "") else "请回答。",
            })
            
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
                
        except Exception as e:
            logger.error(f"Error in Local Qwen3 achat: {e}")
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
        Stream chat with the local Qwen3 model using the given prompt.

        Args:
            prompt: The prompt to chat with.
            history: Conversation history
            **kwargs: Additional arguments

        Yields:
            String chunks of the response.
        """
        try:
            # 延迟加载模型
            self._load_model()
            
            # 对于本地模型，我们模拟流式输出
            # 先生成完整回答，然后分块返回
            answer = self._generate_answer(
                prompt=prompt,
                history=history,
                max_new_tokens=kwargs.get('max_new_tokens', self.max_tokens),
                temperature=kwargs.get('temperature', 0.7),
                top_p=kwargs.get('top_p', 0.9),
                do_sample=kwargs.get('do_sample', True),
                enable_think=kwargs.get('enable_think', False),
            )
            
            # 分块返回（模拟流式）
            chunk_size = 10  # 每次返回10个字符
            for i in range(0, len(answer), chunk_size):
                chunk = answer[i:i + chunk_size]
                yield chunk
                # 添加小延迟以模拟流式效果
                await asyncio.sleep(0.01)
                
        except Exception as e:
            logger.error(f"Error in Local Qwen3 achat_stream: {e}")
            yield "抱歉，处理您的请求时出现了错误。"

    def chat(self, prompt: str, history: list | None = None, **kwargs: Any) -> ModelResponse:
        """
        Synchronous chat with the local Qwen3 model.

        Args:
            prompt: The prompt to chat with.
            history: Conversation history
            **kwargs: Additional arguments

        Returns:
            The response from the model.
        """
        return asyncio.run(self.achat(prompt, history=history, **kwargs))

    def chat_stream(
        self, prompt: str, history: list | None = None, **kwargs: Any
    ) -> Generator[str, None, None]:
        """
        Synchronous stream chat with the local Qwen3 model.

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

    def __del__(self):
        """清理资源"""
        if self._model is not None:
            del self._model
        if self._tokenizer is not None:
            del self._tokenizer
        clear_cuda_cache()
