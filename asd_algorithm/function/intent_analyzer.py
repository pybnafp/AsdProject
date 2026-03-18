import os
import json
from dashscope import Generation
from typing import Dict, Optional


class AliIntentAnalyzer:
    """
    使用阿里云通义千问意图识别模型(tongyi-intent-detect-v3)进行意图分析
    专注于识别自闭症初筛和眼动测试相关意图
    """

    def __init__(self, api_key=None):
        self.api_key = api_key
        # 定义意图字典，仅包含当前需要的两个意图
        self.intent_dict = {
            "screening": "用户想要对自闭症进行",
            "eyetracking": "当用户想要进行眼动相关测试"
        }
        # 构建系统提示词
        self.system_prompt = self._build_system_prompt()

    def _build_system_prompt(self) -> str:
        """构建系统提示词，指导模型进行意图分类"""
        intent_string = json.dumps(self.intent_dict, ensure_ascii=False)
        return f"""你是一个专业的意图分析专家。你的任务是分析用户查询的意图，并结合意图字典定义来回答问题。

意图字典定义：{intent_string}

意图类型说明：
1. screening: 用户明确想要进行自闭症初筛、筛查、评估等专业医学检测，需要具体的筛查工具、问卷或评估流程
2. eyetracking: 用户明确想要进行眼动测试、眼神追踪、视觉注意力等专业眼动相关检测
3. 空字符串: 其他所有类型的查询，包括但不限于：
   - 一般性对话、问候、测试、闲聊等
   - 医学知识咨询（如症状、病因、治疗等）
   - 研究进展、基因、机制等学术问题
   - 抽象问题或概念性问题
   - 不需要专业检测工具的问题

核心判断标准：
- 如果用户明确表达想要进行自闭症初筛、筛查、评估等专业检测，需要具体的筛查工具或问卷 → screening
- 如果用户明确表达想要进行眼动测试、眼神追踪等专业眼动检测 → eyetracking
- 如果是一般性对话、问候、系统测试、简单闲聊、"你好"、"测试一下"、"怎么样"等日常交流内容 → 返回空字符串
- 如果是医学知识咨询（如症状、病因、治疗、基因、研究进展等）→ 返回空字符串
- 如果可以基于已有上下文回答而不需要专业检测工具 → 返回空字符串
- 如果是完全独立、不需要任何外部信息的抽象问题或概念性问题 → 返回空字符串
- 其他所有不匹配screening和eyetracking的查询 → 返回空字符串

重要区分：
- screening: 必须明确表达要进行自闭症初筛/筛查/评估等专业检测，需要具体的检测工具
- eyetracking: 必须明确表达要进行眼动测试/眼神追踪等专业眼动检测
- 空字符串: 医学知识咨询、研究问题、一般对话、抽象问题等所有其他类型

分析要求：
1. 分析query请求时，只有意图结果包含在意图字典定义中的意图时，才显示对应的key，否则就显示为空字符串
2. 严格按照核心判断标准进行意图分类
3. 不要过度推断，如果查询内容不明确，返回空字符串
4. 只返回意图key，不要添加任何额外解释
5. 医学知识咨询和研究问题不属于screening或eyetracking，应返回空字符串

请根据以上标准分析用户查询的意图。"""

    def analyze_intent(self, query: str) -> Dict[str, Optional[str]]:
        """
        分析用户查询的意图

        参数:
            query: 用户输入的查询文本

        返回:
            包含意图分析结果的字典，格式为{"intent": 意图标签或None}
        """
        try:
            # 构建对话消息
            messages = [
                {'role': 'system', 'content': self.system_prompt},
                {'role': 'user', 'content': query}
            ]

            # 调用通义意图识别模型
            response = Generation.call(
                api_key=self.api_key,
                model="tongyi-intent-detect-v3",
                messages=messages,
                result_format="message"
            )

            # 提取意图结果
            intent_result = response.output.choices[0].message.content.strip()
            print(f"intent_result= {intent_result}")
            # 验证结果是否在定义的意图列表中
            if intent_result in self.intent_dict:
                result = {"intent": intent_result}
            else:
                result = {"intent": None}

            # 打印输出以便测试
            print(f"意图分析结果: {json.dumps(result, ensure_ascii=False)}")
            return result

        except Exception as e:
            print(f"意图分析出错: {str(e)}")
            return {"intent": None}


# 测试代码
if __name__ == "__main__":
    # 初始化意图分析器
    api_key = "sk-705575b79f854ea6a25106d794b2b5d2"
    intent_analyzer = AliIntentAnalyzer(api_key)

    # 测试用例
    test_queries = [
        "测试一下",
        "我想要检测自己是否患有自闭症，如何检测？",
        "我想给孩子做自闭症初筛",
        "眼动测试能检测什么问题？",
        "今天天气怎么样？",
        "自闭症筛查需要准备什么材料",
        "请再次测试",
        "眼动追踪技术的原理是什么",
        "孤独症与儿时生活有关吗",
        "孤独症和哪些基因有关",
        "你多大了"
    ]
    # test_queries = [
    #     "I want to get my child screened for autism initially.",
    #     "What problems can eye movement tests detect?",
    #     "What's the weather like today?",
    #     "What materials are needed for autism screening?",
    #     "What is the principle of eye-tracking technology?"
    # ]

    # 执行测试
    for query in test_queries:
        print(f"\n测试查询: {query}")
        intent_analyzer.analyze_intent(query)
