import os

import dashvector
from dashvector import Doc
from dashtext import SparseVectorEncoder

import dashscope
from http import HTTPStatus


class DashVectorRetriever:
    def __init__(self, dashscope_api_key, dashvector_api_key, dashvector_endpoint, collection_name=None):

        self.client = dashvector.Client(api_key=dashvector_api_key, endpoint=dashvector_endpoint)
        self.sparse_encoder = SparseVectorEncoder.default(name='en')
        self.dashscope_api_key = dashscope_api_key
        if collection_name is not None:
            self.collection = self.client.get(name=collection_name)
        else:
            self.collection = None

    def _get_dense_vectors(self, texts, batch_size=5, model="text-embedding-v4", dimension=1024):
        """
        使用dashscope.TextEmbedding生成稠密向量，支持批量
        """
        vectors = []
        for i in range(0, len(texts), batch_size):
            batch = texts[i:i + batch_size]
            resp = dashscope.TextEmbedding.call(
                model=model,
                input=batch,
                dimension=dimension,
                api_key=self.dashscope_api_key
            )
            if resp.status_code == HTTPStatus.OK:
                batch_vectors = [emb['embedding'] for emb in resp.output['embeddings']]
                vectors.extend(batch_vectors)
            else:
                print(f"DashScope embedding error: {resp}")
                # 失败时补零向量
                vectors.extend([[0.0] * dimension for _ in batch])
        return vectors

    def query(self, query_text, top_k):
        # 构造查询向量
        query_dense = self._get_dense_vectors([query_text])[0]
        query_sparse = self.sparse_encoder.encode_documents(query_text)

        # 执行查询
        docs = self.collection.query(
            vector=query_dense,
            sparse_vector=query_sparse,
            topk=top_k
        )
        print(docs)
        results = []
        for i, doc in enumerate(docs.output):
            result = {
                'id': doc.id,
                'score': doc.score,
                'text': doc.fields['text'],
                'metadata': doc.metadata if hasattr(doc, 'metadata') else {}
            }
            results.append(result)
        return results
