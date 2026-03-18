import json
import configparser
import os
import time
import logging
import asyncio
import traceback
from http import HTTPStatus
from pathlib import Path

import pandas as pd
from flask import Flask, request
from dashscope import Generation  # 改用对话生成接口
import dashscope
from function.translate import Translator  # 导入翻译器
from function.intent_analyzer import AliIntentAnalyzer  # 导入意图识别
from function.dashvector_retriever import DashVectorRetriever
from function.generator import Generator
from function.graphrag_retriever import (
    local_graph_search_only_context,
    global_graph_search_only_context,
)
from function.scale_parser import ScaleParser  # 导入量表解析器
from function.multimodal_docx_parser import MultimodalDocxParser  # 导入多模态解析器
import graphrag.api as api
from graphrag.config.load_config import load_config

app = Flask(__name__)
app.json.ensure_ascii = False  # 禁用 ASCII 编码

# 配置 dashscope 日志级别，避免输出非致命性 ERROR 日志
# dashscope SDK 在流式输出时可能会输出一些非致命性的错误日志
# 这些日志不影响功能，但会影响日志可读性
dashscope_logger = logging.getLogger('dashscope')
# 方法1: 设置日志级别为 CRITICAL，只显示严重错误
dashscope_logger.setLevel(logging.CRITICAL)
# 方法2: 如果方法1不够，可以完全禁用日志输出（取消下面的注释）
# dashscope_logger.disabled = True
# 或者移除所有处理器
# dashscope_logger.handlers = []

def print_timing(step_name, elapsed_time):
    """打印耗时信息"""
    print(f"[耗时统计] {step_name}: {elapsed_time:.3f}秒")

def print_separator():
    """打印分隔线"""
    print("=" * 80)

# 读取配置文件信息
config = configparser.ConfigParser()
config.read('config.ini', encoding='utf-8')

# 全局初始化模块
retriever = DashVectorRetriever(config['dashscope']['api_key'], config['dashvector']['api_key'],config['dashvector']['endpoint'], config['dashvector']['collection_name'])  # 阿里的向量服务
translator = Translator(config['dashscope']['api_key'])  # 翻译器
intent_analyzer = AliIntentAnalyzer(config['dashscope']['api_key'])  # 初始化意图分析器
generator = Generator(config['dashscope']['api_key'])

# 初始化量表解析器，传入向量服务配置以支持专业知识检索
scale_parser = ScaleParser(
    config['dashscope']['api_key'],
    config['dashvector']['api_key'],
    config['dashvector']['endpoint'],
    config['dashvector']['collection_name']
)
# 初始化多模态解析器
multimodal_parser = MultimodalDocxParser(
    config['dashscope']['api_key'],
    config['dashvector']['api_key'],
    config['dashvector']['endpoint'],
    config['dashvector']['collection_name']
)

# 初始化GraphRAG配置和数据
GRAPH_PROJECT_DIRECTORY = "./ragtest_test"
graphrag_config = None
graphrag_data = {}

def load_graphrag_data():
    """加载GraphRAG配置和索引数据"""
    global graphrag_config, graphrag_data
    
    try:
        # 加载GraphRAG配置
        project_path = Path(GRAPH_PROJECT_DIRECTORY)
        dashscope_config_path = project_path / "settings_dashscope.yaml"
        
        if not dashscope_config_path.exists():
            print(f"GraphRAG配置文件不存在: {dashscope_config_path}")
            return False
            
        graphrag_config = load_config(project_path, config_filepath=dashscope_config_path)
        print(f"GraphRAG配置加载成功: {dashscope_config_path}")
        
        # 加载索引数据
        output_dir = project_path / "output"
        if not output_dir.exists():
            print(f"GraphRAG输出目录不存在: {output_dir}")
            return False
            
        graphrag_data = {
            'entities': pd.read_parquet(output_dir / "entities.parquet"),
            'communities': pd.read_parquet(output_dir / "communities.parquet"),
            'community_reports': pd.read_parquet(output_dir / "community_reports.parquet"),
            'text_units': pd.read_parquet(output_dir / "text_units.parquet"),
            'relationships': pd.read_parquet(output_dir / "relationships.parquet"),
            'covariates': None  # 可选，如果存在的话
        }
        
        # 检查covariates文件是否存在
        covariates_file = output_dir / "covariates.parquet"
        if covariates_file.exists():
            graphrag_data['covariates'] = pd.read_parquet(covariates_file)
            
        print(f"GraphRAG数据加载成功:")
        print(f"  - 实体: {len(graphrag_data['entities'])} 条")
        print(f"  - 社区: {len(graphrag_data['communities'])} 条")
        print(f"  - 社区报告: {len(graphrag_data['community_reports'])} 条")
        print(f"  - 文本单元: {len(graphrag_data['text_units'])} 条")
        print(f"  - 关系: {len(graphrag_data['relationships'])} 条")
        if graphrag_data['covariates'] is not None:
            print(f"  - 协变量: {len(graphrag_data['covariates'])} 条")
            
        return True
        
    except Exception as e:
        print(f"GraphRAG数据加载失败: {e}")
        return False

# 在应用启动时加载GraphRAG数据
load_graphrag_data()


# 向量服务(阿里向量服务)
@app.route('/api/dashvector_search', methods=['POST', 'GET'])
def dashvector_search():
    print_separator()
    print(f"[API调用] /api/dashvector")
    total_start = time.time()
    
    query_start = time.time()
    retrieval_results = retriever.query(request.values.get("content"),top_k=config.getint('dashvector', 'top_k'))  # 第二步向量检索
    query_elapsed = time.time() - query_start
    print_timing("向量检索", query_elapsed)
    
    total_elapsed = time.time() - total_start
    print_timing("总耗时", total_elapsed)
    print_separator()
    
    return {'success': True, 'result': retrieval_results}, 200


# 大语言模型(阿里算法)
@app.route('/api/dashvector_chat', methods=['POST', 'GET'])
def dashvector_chat():
    print_separator()
    print(f"[API调用] /api/chat")
    total_start = time.time()
    
    # 优先读取 JSON body，再兼容 form/query 参数
    data = request.get_json(silent=True) or {}
    content = data.get("content") or request.values.get("content", "")
    history = data.get("history") or request.values.get("history", "")
    # 防御：如果没有内容，直接返回空结果，避免后续 None 报错
    if not content or not str(content).strip():
        total_elapsed = time.time() - total_start
        print_timing("总耗时", total_elapsed)
        print_separator()
        return {
            "success": True,
            "result": {
                "intention": None,
                "answer": ""
            }
        }, 200

    # 意图识别
    intent_start = time.time()
    intent_result = intent_analyzer.analyze_intent(content)  # 第零步识别意图（中文英文都可识别）
    intent_elapsed = time.time() - intent_start
    print_timing("意图识别", intent_elapsed)
    
    # 翻译
    translate_start = time.time()
    processed_query = translator.preprocess_query(content)  # 第一步先翻译
    translate_elapsed = time.time() - translate_start
    print_timing("翻译处理", translate_elapsed)
    
    # 向量检索
    retrieval_start = time.time()
    retrieval_results = retriever.query(processed_query, top_k=3)  # 第二步向量检索
    retrieval_elapsed = time.time() - retrieval_start
    print_timing("向量检索", retrieval_elapsed)
    
    # 准备参考资料
    prepare_start = time.time()
    filtered_content = generator.prepare_reference_content(retrieval_results)
    user_message = generator.get_prompt(filtered_content, content)
    prepare_elapsed = time.time() - prepare_start
    print_timing("内容准备", prepare_elapsed)
    
    # 组装提示词
    message_start = time.time()
    messages = [{
        "role": "system",
        "content": generator.system_message
    }]
    if not history or not history.strip() or history == "null":
        print('没有历史记录')
    else:
        history_json = json.loads(history)
        for json_ in history_json:
            messages.append({"role": "user", "content": json_["prompt"]})
            messages.append({"role": "assistant", "content": json_["completion"]})
    messages.append({"role": "user", "content": user_message})
    message_elapsed = time.time() - message_start
    print_timing("消息组装", message_elapsed)
    
    # 调用大语言模型（流式输出）
    llm_start = time.time()
    print(messages)
    content_parts = []
    usage = None
    
    try:
        responses = Generation.call(  # 第三步调用阿里算法
            api_key=config['dashscope']['api_key'],
            model=config['dashscope']['model'],
            messages=messages,
            result_format="message",
            max_tokens=config['dashscope']['max_tokens'],
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
                return {'success': False, 'error': error_msg}, 500
        
        answer = "".join(content_parts)
        llm_elapsed = time.time() - llm_start
        print_timing("大语言模型生成", llm_elapsed)
        
        if usage:
            print("--- 请求用量 ---")
            print(f"输入 Tokens: {usage.input_tokens}")
            print(f"输出 Tokens: {usage.output_tokens}")
            print(f"总计 Tokens: {usage.total_tokens}")
    
    except Exception as e:
        llm_elapsed = time.time() - llm_start
        print_timing("大语言模型生成", llm_elapsed)
        error_msg = f"大语言模型调用出错: {str(e)}"
        print(error_msg)
        return {'success': False, 'error': error_msg}, 500

    # 整合意图分析结果和回答
    final_result = {
        "intention": intent_result["intent"],  # 意图分析结果
        "answer": answer  # AI生成的回答
    }
    
    total_elapsed = time.time() - total_start
    print_timing("总耗时", total_elapsed)
    print_separator()
    
    return {'success': True, 'result': final_result}, 200


# GraphRAG本地搜索接口（graph检索 + 通义 Generation.call）
@app.route('/api/local_graph_chat', methods=['POST', 'GET'])
def local_graph_chat():
    """GraphRAG本地搜索接口：先进行graph检索，再用通义大模型生成回答。"""
    try:
        # 检查GraphRAG数据是否已加载
        if graphrag_config is None or not graphrag_data:
            return {
                'success': False,
                'error': 'GraphRAG数据未加载，请检查配置文件和索引数据'
            }, 500

        # 获取查询内容
        content = request.values.get("content")
        if not content or not content.strip():
            return {
                'success': False,
                'error': '查询内容不能为空'
            }, 400

        print(f"查询内容: {content}")
        print("开始执行GraphRAG本地搜索（graph检索 + 通义LLM回答）...")

        # 第一步：执行本地图检索，获取graph上下文
        async def run_local_graph_search():
            return await local_graph_search_only_context(
                config=graphrag_config,
                entities=graphrag_data['entities'],
                communities=graphrag_data['communities'],
                community_reports=graphrag_data['community_reports'],
                text_units=graphrag_data['text_units'],
                relationships=graphrag_data['relationships'],
                covariates=graphrag_data['covariates'],
                community_level=2,
                response_type="Multiple Paragraphs",
                verbose=False,
                query=content,
            )

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            context_result = loop.run_until_complete(run_local_graph_search())
        finally:
            loop.close()

        context_chunks = context_result.get("context_chunks", "")

        # 第二步：组装提示词，调用通义 Generation
        # 将graph检索结果作为参考资料注入到用户消息中
        user_message = (
            "下面是基于知识图谱本地搜索得到的检索结果，请严格以这些内容为主要依据，"
            "结合必要的一般常识，用中文客观、详尽地回答用户问题。\n\n"
            f"【用户问题】\n{content}\n\n"
            f"【本地GraphRAG检索结果】\n{context_chunks}"
        )

        messages = [{
            "role": "system",
            "content": generator.system_message
        }, {
            "role": "user",
            "content": user_message
        }]

        llm_start = time.time()
        content_parts = []
        usage = None

        try:
            responses = Generation.call(
                api_key=config['dashscope']['api_key'],
                model=config['dashscope']['model'],
                messages=messages,
                result_format="message",
                max_tokens=config['dashscope']['max_tokens'],
                stream=True,
                incremental_output=True,
            )

            for resp in responses:
                if resp.status_code == HTTPStatus.OK:
                    chunk = resp.output.choices[0].message.content
                    content_parts.append(chunk)
                    if resp.output.choices[0].finish_reason == "stop":
                        usage = resp.usage
                        break
                else:
                    error_msg = (
                        f"请求失败: request_id={resp.request_id}, "
                        f"code={resp.code}, message={resp.message}"
                    )
                    print(error_msg)
                    return {'success': False, 'error': error_msg}, 500

            answer = "".join(content_parts)
            llm_elapsed = time.time() - llm_start
            print_timing("本地search大语言模型生成", llm_elapsed)

            if usage:
                print("--- 请求用量 ---")
                print(f"输入 Tokens: {usage.input_tokens}")
                print(f"输出 Tokens: {usage.output_tokens}")
                print(f"总计 Tokens: {usage.total_tokens}")

        except Exception as e:
            llm_elapsed = time.time() - llm_start
            print_timing("本地search大语言模型生成", llm_elapsed)
            error_msg = f'GraphRAG本地搜索LLM调用出错: {str(e)}'
            print(error_msg)
            return {'success': False, 'error': error_msg}, 500

        final_result = {
            "query": content,
            "response": answer,
            "graph_context": {
                "context_chunks": context_chunks,
                "context_records": context_result.get("context_records", {}),
            },
        }

        return {
            'success': True,
            'result': final_result
        }, 200

    except Exception as e:
        print(f"GraphRAG本地搜索失败: {e}")
        traceback.print_exc()
        return {
            'success': False,
            'error': f'GraphRAG本地搜索失败：{str(e)}'
        }, 500


# GraphRAG本地图检索接口（仅返回graph检索结果，不调用LLM）
@app.route('/api/local_graph_search', methods=['POST', 'GET'])
def local_graph_search():
    """GraphRAG本地图检索接口，仅返回图检索结果（entities/relationships/sources/reports等），不调用LLM。"""
    try:
        # 检查GraphRAG数据是否已加载
        if graphrag_config is None or not graphrag_data:
            return {
                'success': False,
                'error': 'GraphRAG数据未加载，请检查配置文件和索引数据'
            }, 500

        # 获取查询内容
        content = request.values.get("content")
        if not content or not content.strip():
            return {
                'success': False,
                'error': '查询内容不能为空'
            }, 400

        print(f"查询内容: {content}")
        print("开始执行GraphRAG本地图检索（仅返回graph上下文，不调用LLM）...")

        async def run_local_graph_search():
            return await local_graph_search_only_context(
                config=graphrag_config,
                entities=graphrag_data['entities'],
                communities=graphrag_data['communities'],
                community_reports=graphrag_data['community_reports'],
                text_units=graphrag_data['text_units'],
                relationships=graphrag_data['relationships'],
                covariates=graphrag_data['covariates'],
                community_level=2,
                response_type="Multiple Paragraphs",
                verbose=False,
                query=content,
            )

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            context_result = loop.run_until_complete(run_local_graph_search())
        finally:
            loop.close()

        print("GraphRAG本地图检索完成（未调用LLM）。")

        final_result = {
            "query": content,
            "context_chunks": context_result.get("context_chunks", ""),
            "context_records": context_result.get("context_records", {}),
        }

        return {
            'success': True,
            'result': final_result
        }, 200

    except Exception as e:
        print(f"GraphRAG本地图检索失败: {e}")
        traceback.print_exc()
        return {
            'success': False,
            'error': f'GraphRAG本地图检索失败：{str(e)}'
        }, 500


# GraphRAG全局搜索接口（graph检索 + Map-Reduce 两阶段 + 通义 Generation.call）
@app.route('/api/global_graph_chat', methods=['POST', 'GET'])
def global_graph_chat():
    """GraphRAG全局搜索接口：先全局graph检索，再用Map-Reduce两阶段调用通义大模型生成回答。"""
    try:
        # 检查GraphRAG数据是否已加载
        if graphrag_config is None or not graphrag_data:
            return {
                'success': False,
                'error': 'GraphRAG数据未加载，请检查配置文件和索引数据'
            }, 500

        # 获取查询内容
        content = request.values.get("content")
        if not content or not content.strip():
            return {
                'success': False,
                'error': '查询内容不能为空'
            }, 400

        print(f"查询内容: {content}")
        print("开始执行GraphRAG全局搜索（graph检索 + Map-Reduce + 通义LLM回答）...")

        # 第一步：执行全局图检索，获取社区报告上下文
        async def run_global_graph_search():
            return await global_graph_search_only_context(
                config=graphrag_config,
                entities=graphrag_data['entities'],
                communities=graphrag_data['communities'],
                community_reports=graphrag_data['community_reports'],
                community_level=2,
                dynamic_community_selection=False,
                response_type="Multiple Paragraphs",
                verbose=False,
                query=content,
            )

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            context_result = loop.run_until_complete(run_global_graph_search())
        finally:
            loop.close()

        context_chunks = context_result.get("context_chunks", [])
        if not isinstance(context_chunks, list):
            # 兼容异常情况
            context_chunks = [str(context_chunks)]

        # 第二步：Map 阶段，对每个社区批次调用一次通义LLM生成局部回答
        map_answers = []
        for idx, chunk in enumerate(context_chunks):
            map_prompt = (
                "下面是一部分全局社区报告内容，请基于这段内容，"
                "用中文简要回答用户问题，给出与本段内容相关的要点小结，"
                "不要超过300字，避免重复赘述。\n\n"
                f"【用户问题】\n{content}\n\n"
                f"【社区报告分片 #{idx + 1}】\n{chunk}"
            )

            messages = [{
                "role": "system",
                "content": generator.system_message
            }, {
                "role": "user",
                "content": map_prompt
            }]

            try:
                resp = Generation.call(
                    api_key=config['dashscope']['api_key'],
                    model=config['dashscope']['model'],
                    messages=messages,
                    result_format="message",
                    max_tokens=config['dashscope']['max_tokens'],
                    stream=False,
                )
                if resp.status_code == HTTPStatus.OK:
                    ans = resp.output.choices[0].message.content
                    map_answers.append(ans)
                else:
                    error_msg = (
                        f"全局search Map阶段请求失败: request_id={resp.request_id}, "
                        f"code={resp.code}, message={resp.message}"
                    )
                    print(error_msg)
            except Exception as e:
                print(f"全局search Map阶段调用异常: {e}")

        if not map_answers:
            return {
                'success': False,
                'error': '全局graph检索或Map阶段未能生成任何结果'
            }, 500

        # 第三步：Reduce 阶段，将所有分批小结汇总成最终回答
        reduce_prompt = (
            "下面是若干批次社区报告的小结，请你综合这些小结内容，"
            "用中文给出一个结构化、条理清晰的最终回答，重点突出关键信息，"
            "避免机械重复，必要时可以适当补充一般常识，但不要胡编乱造。\n\n"
            f"【用户问题】\n{content}\n\n"
            "【社区报告分批小结】\n"
        )
        for i, ans in enumerate(map_answers, 1):
            reduce_prompt += f"--- 小结 #{i} ---\n{ans}\n\n"

        messages = [{
            "role": "system",
            "content": generator.system_message
        }, {
            "role": "user",
            "content": reduce_prompt
        }]

        try:
            resp = Generation.call(
                api_key=config['dashscope']['api_key'],
                model=config['dashscope']['model'],
                messages=messages,
                result_format="message",
                max_tokens=config['dashscope']['max_tokens'],
                stream=False,
            )
            if resp.status_code != HTTPStatus.OK:
                error_msg = (
                    f"全局search Reduce阶段请求失败: request_id={resp.request_id}, "
                    f"code={resp.code}, message={resp.message}"
                )
                print(error_msg)
                return {'success': False, 'error': error_msg}, 500
            final_answer = resp.output.choices[0].message.content
        except Exception as e:
            error_msg = f'GraphRAG全局搜索Reduce阶段LLM调用出错: {str(e)}'
            print(error_msg)
            return {'success': False, 'error': error_msg}, 500

        final_result = {
            "query": content,
            "response": final_answer,
            "graph_context": {
                "context_chunks": context_chunks,
                "context_records": context_result.get("context_records", {}),
            },
            "map_answers": map_answers,
        }

        return {
            'success': True,
            'result': final_result
        }, 200

    except Exception as e:
        print(f"GraphRAG全局搜索失败: {e}")
        traceback.print_exc()
        return {
            'success': False,
            'error': f'GraphRAG全局搜索失败：{str(e)}'
        }, 500


# GraphRAG全局图检索接口（仅返回graph检索结果，不调用LLM）
@app.route('/api/global_graph_search', methods=['POST', 'GET'])
def global_graph_search():
    """GraphRAG全局图检索接口，仅返回基于社区报告的全局graph检索结果，不调用LLM。"""
    try:
        # 检查GraphRAG数据是否已加载
        if graphrag_config is None or not graphrag_data:
            return {
                'success': False,
                'error': 'GraphRAG数据未加载，请检查配置文件和索引数据'
            }, 500

        # 获取查询内容
        content = request.values.get("content")
        if not content or not content.strip():
            return {
                'success': False,
                'error': '查询内容不能为空'
            }, 400

        print(f"查询内容: {content}")
        print("开始执行GraphRAG全局图检索（仅返回graph上下文，不调用LLM）...")

        async def run_global_graph_search():
            return await global_graph_search_only_context(
                config=graphrag_config,
                entities=graphrag_data['entities'],
                communities=graphrag_data['communities'],
                community_reports=graphrag_data['community_reports'],
                community_level=2,
                dynamic_community_selection=False,
                response_type="Multiple Paragraphs",
                verbose=False,
                query=content,
            )

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            context_result = loop.run_until_complete(run_global_graph_search())
        finally:
            loop.close()

        print("GraphRAG全局图检索完成（未调用LLM）。")

        final_result = {
            "query": content,
            "context_chunks": context_result.get("context_chunks", []),
            "context_records": context_result.get("context_records", {}),
        }

        return {
            'success': True,
            'result': final_result
        }, 200

    except Exception as e:
        print(f"GraphRAG全局图检索失败: {e}")
        traceback.print_exc()
        return {
            'success': False,
            'error': f'GraphRAG全局图检索失败：{str(e)}'
        }, 500


@app.route('/api/scale_analysis', methods=['POST', 'GET'])
def scale_analysis():
    """量表内容解析接口"""
    print_separator()
    print(f"[API调用] /api/scale_analysis")
    total_start = time.time()
    
    # 获取请求参数
    content = request.values.get("content", "")  # 量表数据

    # 创建量表JSON数据
    parse_start = time.time()
    scale_data = scale_parser.create_scale_json(content)
    parse_elapsed = time.time() - parse_start
    print_timing("量表数据解析", parse_elapsed)
    
    # 使用向量检索增强的AI分析（100字限制）
    analysis_start = time.time()
    analysis_result = scale_parser.analyze_scale_with_ai(scale_data)
    analysis_elapsed = time.time() - analysis_start
    print_timing("AI分析处理", analysis_elapsed)

    total_elapsed = time.time() - total_start
    print_timing("总耗时", total_elapsed)
    print_separator()
    
    return {'success': True, 'result': analysis_result}, 200

@app.route('/api/multimodal_eye_analysis', methods=['POST', 'GET'])
def multimodal_eye_analysis():
    """多模态眼动测试内容解析接口"""
    print_separator()
    print(f"[API调用] /api/multimodal_eye_analysis")
    total_start = time.time()
    
    file_path = None
    fmri_output_folder = None

    # 获取上传的文件
    file_start = time.time()
    fmri_file = request.files.get('fmri')
    if fmri_file and fmri_file.filename != '':
        # fmri = True
        fmri_output_folder = os.path.join('./data/' + request.values.get("path", "default"), 'fmri')
        if not os.path.exists(fmri_output_folder):
            os.makedirs(fmri_output_folder)

        # 安全文件名处理
        from werkzeug.utils import secure_filename
        file_path = os.path.join(fmri_output_folder, secure_filename(fmri_file.filename))
        fmri_file.save(file_path)
        file_elapsed = time.time() - file_start
        print_timing("文件保存", file_elapsed)

        # 使用多模态解析器处理文件
        parse_start = time.time()
        eye_data = multimodal_parser.process_eye_file(file_path, config['dashscope']['api_key'])
        parse_elapsed = time.time() - parse_start
        print_timing("多模态解析处理", parse_elapsed)

        if "error" in eye_data:
            total_elapsed = time.time() - total_start
            print_timing("总耗时", total_elapsed)
            print_separator()
            return {
                'success': False,
                'error': eye_data["error"]
            }, 500

        # 清理临时文件和文件夹
        cleanup_start = time.time()
        if file_path and os.path.exists(file_path):
            os.remove(file_path)
            print(f"临时文件已删除: {file_path}")
        # 删除文件夹（如果为空）
        if fmri_output_folder and os.path.exists(fmri_output_folder):
            os.rmdir(fmri_output_folder)  # 只能删除空文件夹
            print(f"临时文件夹已删除: {fmri_output_folder}")
        cleanup_elapsed = time.time() - cleanup_start
        print_timing("文件清理", cleanup_elapsed)

        total_elapsed = time.time() - total_start
        print_timing("总耗时", total_elapsed)
        print_separator()
        
        # 返回多模态分析结果
        return {
            'success': True,
            'result': eye_data["analysis_result"],
            'document_parts': eye_data["document_parts"]
        }, 200
    
    total_elapsed = time.time() - total_start
    print_timing("总耗时", total_elapsed)
    print_separator()
    
    return {'success': False, 'error': '未上传文件'}, 400


if __name__ == "__main__":
    app.run(host='0.0.0.0', port=5003)
