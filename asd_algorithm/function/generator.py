import os
from dashscope import Generation


class Generator:
    def __init__(self, api_key=None, model="qwen-plus"):
        """
        初始化通义千问生成模型
        Args:
            api_key: 阿里云百炼API Key，如果为None则从环境变量DASHSCOPE_API_KEY获取
            model: 模型名称，默认为qwen-plus
        """
        self.api_key = api_key
        if not self.api_key:
            raise ValueError("请提供DASHSCOPE_API_KEY环境变量或直接传入api_key参数")
        self.model = model
        # 系统提示词
        self.system_message = """你是一个专业的医疗问答助手，专门负责各种医学症状相关的咨询。
                                请给出专业的回答，包括简要总述和分点回答
                                要求回答要准确、专业、有用。"""

    def prepare_reference_content(self, context_docs):
        """
        准备参考资料内容，重构文档结构
        Args:
            context_docs: 检索到的相关文档列表
        Returns:
            格式化的参考资料内容
        """
        try:
            content = ""
            for doc in context_docs:
                if isinstance(doc, dict):
                    # 处理文档引用
                    citation_id = self.process_document_citation(doc)
                    doc_text = doc.get('text', '')
                    content += f"\n[文件: {citation_id}]\n{doc_text}\n"
                else:
                    # 如果不是字典格式，直接添加
                    content += f"\n[文件: 未知来源]\n{str(doc)}\n"

            return content.strip()

        except Exception as e:
            print(f"准备参考资料内容时出错: {e}")
            return ""

    def process_document_citation(self, doc):
        """
        处理文档引用，提取PMID或其他标识符
        Args:
            doc: 文档字典
        Returns:
            引用标识符
        """
        try:
            # 尝试从文档ID中提取PMID
            if 'id' in doc:
                doc_id = doc['id']
                # 从ID中提取PMID号（如 "32716138_39" 中的 "32716138"）
                if '_' in doc_id:
                    pmid = doc_id.split('_')[0]
                    return f"PubMed: {pmid}"
                else:
                    return f"PubMed: {doc_id}"

            # 如果没有ID，尝试从metadata中获取
            if 'metadata' in doc and isinstance(doc['metadata'], dict):
                metadata = doc['metadata']
                if 'source' in metadata:
                    return f"来源: {metadata['source']}"
                elif 'title' in metadata:
                    return f"标题: {metadata['title']}"

            return "未知来源"

        except Exception as e:
            print(f"处理文档引用时出错: {e}")
            return "未知来源"

    def get_prompt(self, filtered_content, message):
        """
        构建用户提示词
        Args:
            filtered_content: 格式化的参考资料内容
            message: 用户问题
        Returns:
            完整的用户提示词
        """
        return f"""你是 StellarCare AI，一个专门服务于有孤独症谱系障碍(ASD)儿童家庭的AI助手。现在，请基于检索到的专业参考资料，为家庭提供专业、温暖的支持。

        参考资料内容：
        ---------------------
        {filtered_content}
        ---------------------

        回答指南：
        1. 认真分析用户的问题，提供准确、实用的解决方案
        2. 引用参考资料时使用"据[PubMed: filename](https://pubmed.ncbi.nlm.nih.gov/[filename])"格式
        3. 如果参考资料不足，诚实说明并提供基于专业知识的建议
        4. 使用温暖、易懂的语言传达专业信息
        5. 在回答问题的同时，给予适当的情感支持
        6. 必要时建议咨询专业医疗团队
        7. 请按以下格式回答：
           1. [第一个要点]
           2. [第二个要点]
           3. [第三个要点]
           ...（根据内容需要添加更多要点）

        **用户的当前问题：**
        {message}"""
