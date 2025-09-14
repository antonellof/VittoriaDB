#!/usr/bin/env python3
"""
Test all external vectorizer services (clean implementation)
"""

import sys
import os
import vittoriadb
from vittoriadb.configure import Configure
import time

def test_vectorizer(name, config, expected_status="needs_setup"):
    print(f"\n🧪 TESTING {name.upper()}")
    print("=" * 60)
    
    client = vittoriadb.connect(url="http://localhost:8080")
    collection_name = f"Test_{name.replace(' ', '')}_{int(time.time())}"
    
    try:
        collection = client.create_collection(
            name=collection_name,
            dimensions=config.dimensions,
            metric="cosine",
            vectorizer_config=config
        )
        
        print("✅ Collection created successfully")
        
        # Test with a simple document
        test_text = "Artificial intelligence transforms how we understand information processing."
        collection.insert_text("test_doc", test_text)
        print("✅ Document inserted successfully")
        
        # Test search
        results = collection.search_text("machine learning and AI", limit=1)
        if results:
            print(f"✅ Search successful - similarity: {results[0].score:.4f}")
            return {"status": "working", "similarity": results[0].score}
        else:
            print("⚠️  Search returned no results")
            return {"status": "partial", "similarity": 0}
            
    except Exception as e:
        print(f"❌ Failed: {e}")
        return {"status": expected_status, "error": str(e)}

def main():
    print("🧪 COMPREHENSIVE VECTORIZER TEST (External Services Only)")
    print("=" * 80)
    print("Testing all external service vectorizers following industry patterns")
    print()
    
    # Test configurations
    vectorizers = [
        ("Ollama Local", Configure.Vectors.ollama_embeddings(), "needs_ollama"),
        ("auto_embeddings", Configure.Vectors.auto_embeddings(), "needs_ollama"),
        ("Sentence Transformers", Configure.Vectors.sentence_transformers(), "needs_python"),
        ("OpenAI", Configure.Vectors.openai_embeddings(api_key="dummy_key"), "needs_api_key"),
        ("HuggingFace", Configure.Vectors.huggingface_embeddings(api_key="dummy_key"), "needs_api_key"),
    ]
    
    results = {}
    
    for name, config, expected_status in vectorizers:
        result = test_vectorizer(name, config, expected_status)
        results[name] = result
    
    # Summary
    print("\n" + "="*80)
    print("📊 VECTORIZER STATUS SUMMARY")
    print("="*80)
    
    print(f"{'Vectorizer':<20} {'Status':<15} {'Type':<15} {'Dependencies':<20}")
    print("-" * 70)
    
    vectorizer_info = [
        ("Ollama Local", "Local ML Model", "Ollama installation"),
        ("auto_embeddings", "Local ML Model", "Ollama installation"),
        ("Sentence Transformers", "Local ML Process", "Python + models"),
        ("OpenAI", "Cloud API", "API key + credits"),
        ("HuggingFace", "Cloud API", "API token"),
    ]
    
    for i, (name, vtype, deps) in enumerate(vectorizer_info):
        result = results.get(name, {"status": "unknown"})
        status_icon = {
            "working": "✅ Working",
            "partial": "⚠️  Partial", 
            "needs_ollama": "🔧 Need Ollama",
            "needs_python": "🐍 Need Python",
            "needs_api_key": "🔑 Need API Key",
            "unknown": "❓ Unknown"
        }.get(result["status"], "❓ Unknown")
        
        print(f"{name:<20} {status_icon:<15} {vtype:<15} {deps:<20}")
    
    print(f"\n🎯 SETUP INSTRUCTIONS:")
    print(f"")
    print(f"🔧 For Ollama (Recommended for local ML):")
    print(f"   1. Install: curl -fsSL https://ollama.ai/install.sh | sh")
    print(f"   2. Start: ollama serve")
    print(f"   3. Pull model: ollama pull nomic-embed-text")
    print(f"   4. Use: Configure.Vectors.auto_embeddings()")
    print(f"")
    print(f"🐍 For Sentence Transformers:")
    print(f"   1. Install: pip install sentence-transformers")
    print(f"   2. Use: Configure.Vectors.sentence_transformers()")
    print(f"")
    print(f"🔑 For OpenAI:")
    print(f"   1. Get API key: https://platform.openai.com/api-keys")
    print(f"   2. Use: Configure.Vectors.openai_embeddings(api_key='your_key')")
    print(f"")
    print(f"🤗 For HuggingFace:")
    print(f"   1. Get token: https://huggingface.co/settings/tokens")
    print(f"   2. Use: Configure.Vectors.huggingface_embeddings(api_key='your_token')")
    
    print(f"\n🏆 RECOMMENDATIONS:")
    print(f"   🥇 Best Overall: Ollama (high quality + local + no costs)")
    print(f"   🥈 Best Cloud: OpenAI (highest quality, costs money)")
    print(f"   🥉 Best Free: HuggingFace (good quality, free tier)")
    print(f"   🔬 Best Control: Sentence Transformers (full local control)")

if __name__ == "__main__":
    main()
