import json
from typing import Dict, List, Any
from http import HTTPStatus
from dashscope import Generation
from function.dashvector_retriever import DashVectorRetriever


class ScaleParser:
    """儿童心理行为发育问题预警征象筛查表解析器"""

    def __init__(
        self,
        api_key: str,
        vector_api_key: str = None,
        vector_endpoint: str = None,
        collection_name: str = None,
        retriever: DashVectorRetriever = None,
    ):
        self.api_key = api_key
        # 初始化向量检索器（如果提供了向量服务配置）
        self.vector_retriever = retriever
        if self.vector_retriever is None and vector_api_key and vector_endpoint and collection_name:
            self.vector_retriever = DashVectorRetriever(api_key, vector_api_key, vector_endpoint, collection_name)

    def create_scale_json(self, content: str) -> Dict[str, Any]:
        """创建儿童心理行为发育问题预警征象筛查表的JSON结构"""
        content = json.loads(content)
        print(content)
        scale_data = {
            "scale_name": "问卷量表分析",
            "scale_description": "分析不同月龄阶段的儿童发育问题",
            "gender": content["gender"],
            "birth_date": content["birth"],
            "age_groups": content["items"],
        }
        # print(scale_data)

        return scale_data

    def analyze_scale_with_ai(self, scale_data: Dict[str, Any]) -> str:
        """使用通义模型分析量表内容，集成向量检索提高专业性"""

        # 构建系统提示词
        system_prompt = """你是一个温和专业的儿童发育评估专家。请根据提供的儿童心理行为发育问题预警征象筛查表内容，进行温和、鼓励性的分析总结。

重要要求：
1. 回答必须控制在100字左右
2. 使用温和、鼓励的语气，避免过于严肃或吓人的表述
3. 强调这是预警性筛查，不是诊断，重在早期发现和积极干预
4. 语言亲切易懂，适合家长理解
5. 针对异常项，从以下四个方面进行温和分析：
   - 观察要点：温和地说明需要关注的发育表现
   - 发展特点：以发展的角度看待个体差异
   - 支持建议：提供积极、可操作的亲子互动建议
   - 关注重点：说明后续观察的要点，保持积极态度

请用温暖、专业的语调，给家长以信心和支持，强调每个孩子都有自己的发展节奏。"""

        # 构建用户消息
        user_query = "请温和地分析这个儿童的发育情况，给家长一些积极的建议。"

        # 分析量表数据，判断正常/异常
        assessment_result = self._analyze_scale_items(scale_data)
        print("量表判断正常/异常 assessment_result=", assessment_result)
        # 向量检索相关专业知识
        retrieved_knowledge = ""
        if self.vector_retriever:
            try:
                # 构建检索查询，结合量表内容
                search_query = ""
                for group in scale_data["age_groups"]:
                    search_query += group["question"] + " "
                search_query = f"{search_query} 儿童发育评估"
                retrieval_results = self.vector_retriever.query(search_query, top_k=2)

                if retrieval_results and len(retrieval_results) > 0:
                    retrieved_knowledge = "\n\n专业知识参考：\n"
                    for i, result in enumerate(retrieval_results):
                        # 截取关键信息，避免过长
                        # text = result['text'][:200] + "..." if len(result['text']) > 200 else result['text']
                        retrieved_knowledge += f"{i}. {result['text']}\n"
                    retrieved_knowledge += "\n"
                print("retrieved_knowledge=", retrieved_knowledge)
            except Exception as e:
                print(f"向量检索过程中出现错误：{str(e)}")
                retrieved_knowledge = "\n\n注意：无法获取相关专业知识参考。\n"

        user_message = f"""
请温和地分析以下儿童发育筛查数据：

{json.dumps(scale_data, ensure_ascii=False, indent=2)}

筛查结果：{assessment_result['status']}
需要关注的方面：{assessment_result['abnormal_items'] if assessment_result['status'] == '异常' else '发育良好'}

用户问题：{user_query}{retrieved_knowledge}

请用温暖、鼓励的语气，在100字内给出积极的分析建议，给家长信心和支持。
"""

        # 调用通义模型（流式输出）
        try:
            content_parts = []
            usage = None
            
            responses = Generation.call(
                api_key=self.api_key,
                model='qwen-plus',
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_message}
                ],
                result_format="message",
                max_tokens=150,
                stream=True,
                incremental_output=True  # 关键：设置为True以获取增量输出，性能更佳
            )
            
            for resp in responses:
                if resp.status_code == HTTPStatus.OK:
                    content = resp.output.choices[0].message.content
                    content_parts.append(content)
                    
                    # 检查是否是最后一个包
                    if resp.output.choices[0].finish_reason == "stop":
                        usage = resp.usage
                        break
                else:
                    # 处理错误情况
                    error_msg = f"请求失败: request_id={resp.request_id}, code={resp.code}, message={resp.message}"
                    print(error_msg)
                    return f"分析过程中出现错误：{error_msg}"
            
            result = "".join(content_parts)
            
            if usage:
                print("--- 请求用量 ---")
                print(f"输入 Tokens: {usage.input_tokens}")
                print(f"输出 Tokens: {usage.output_tokens}")
                print(f"总计 Tokens: {usage.total_tokens}")
            
            return result

        except Exception as e:
            return f"分析过程中出现错误：{str(e)}"

    def _analyze_scale_items(self, scale_data: Dict[str, Any]) -> Dict[str, Any]:
        """分析量表项目，判断正常/异常状态"""
        abnormal_items = []
        all_items = scale_data.get("age_groups", [])

        for item in all_items:
            warning_sign = item.get("warning_sign", "")
            response_options = item.get("response_options", "")

            # 判断是否为异常项
            # 如果response_options为"negative"或False，表示存在预警征象
            if response_options in ["negative", False, "无反应", "不会", "不能", "不会微笑", "不会发声", "不会抬头"]:
                abnormal_items.append({
                    "warning_sign": warning_sign,
                    "dimension": item.get("dimension", ""),
                    "question": item.get("question", "")
                })

        # 判断整体状态
        if len(abnormal_items) == 0:
            status = "正常"
        else:
            status = "异常"

        return {
            "status": status,
            "abnormal_items": abnormal_items,
            "total_items": len(all_items),
            "abnormal_count": len(abnormal_items)
        }
