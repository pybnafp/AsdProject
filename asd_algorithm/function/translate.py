# preprocess/translate.py
import os
import re
from typing import Tuple
from dashscope import Generation  # 改用对话生成接口


class Translator:
    def __init__(self, api_key=None):
        self.api_key = api_key

    def is_chinese(self, text: str) -> bool:
        chinese_pattern = re.compile(r'[\u4e00-\u9fff]')
        return bool(chinese_pattern.search(text))

    def translate_to_english(self, text):
        prompt = f"将以下中文翻译成英文，仅返回翻译结果，不要额外内容：\n{text}"
        response = Generation.call(
            model="qwen-plus",  # 使用qwen-plus模型
            prompt=prompt,
            api_key=self.api_key
        )
        if response.status_code == 200:
            return response.output.text.strip()
        else:
            raise Exception(f"翻译失败：{response.message}")

    # def translate_to_chinese(self, text):
    #     """将英文翻译成中文"""
    #     prompt = f"将以下英文翻译成中文，仅返回翻译结果，不要额外内容：\n{text}"
    #     response = Generation.call(
    #         model="qwen-plus",  # 使用qwen-plus模型
    #         prompt=prompt,
    #         api_key=self.api_key
    #     )
    #     if response.status_code == 200:
    #         return response.output.text.strip()
    #     else:
    #         raise Exception(f"翻译失败：{response.message}")

    def preprocess_query(self, query: str) -> Tuple[str, str]:
        if self.is_chinese(query):
            translated = self.translate_to_english(query)
            return translated
        else:
            return ''
