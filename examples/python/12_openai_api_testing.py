#!/usr/bin/env python3
"""
Test OpenAI vectorizer with real API key
"""

import sys
import os
import vittoriadb
from vittoriadb.configure import Configure
import time

def test_openai_with_key():
    print("🤖 TESTING OPENAI WITH API KEY")
    print("=" * 60)
    
    # Method 1: Set your API key directly in code (for testing)
    # REPLACE 'your_api_key_here' with your actual OpenAI API key
    api_key = "your_api_key_here"  # ⚠️ Replace this!
    
    # Method 2: Read from environment variable (recommended)
    # export OPENAI_API_KEY="your_actual_key"
    if api_key == "your_api_key_here":
        api_key = os.getenv("OPENAI_API_KEY")
        if not api_key:
            print("❌ No OpenAI API key found!")
            print("\n💡 To test with OpenAI, you can:")
            print("   Option 1: Set environment variable:")
            print("     export OPENAI_API_KEY='your_actual_key'")
            print("     python3 test_openai_with_key.py")
            print("")
            print("   Option 2: Edit this file and replace 'your_api_key_here' with your key")
            print("")
            print("   Option 3: Pass key directly:")
            print("     Configure.Vectors.openai_embeddings(api_key='your_key')")
            return
    
    client = vittoriadb.connect(url="http://localhost:8080")
    
    try:
        collection_name = f"OpenAITest_{int(time.time())}"
        collection = client.create_collection(
            name=collection_name,
            dimensions=1536,  # OpenAI text-embedding-ada-002 dimensions
            metric="cosine",
            vectorizer_config=Configure.Vectors.openai_embeddings(
                model="text-embedding-ada-002",
                api_key=api_key
            )
        )
        
        print("✅ Collection created with OpenAI embeddings")
        
        # Test with diverse content
        test_docs = {
            "technology": "Artificial intelligence and machine learning algorithms",
            "science": "Quantum physics and molecular chemistry research",
            "literature": "Creative writing and poetic expression",
            "medicine": "Healthcare treatments and medical diagnostics",
            "philosophy": "Consciousness and existential questions"
        }
        
        print("\n📝 Inserting documents via OpenAI API...")
        for doc_id, text in test_docs.items():
            collection.insert_text(doc_id, text)
            print(f"   ✅ {doc_id}: {text}")
        
        print("\n🔍 Testing semantic search...")
        
        queries = [
            ("AI query", "machine learning and algorithms", "technology"),
            ("Science query", "physics and chemistry", "science"),
            ("Creative query", "writing and poetry", "literature"),
            ("Health query", "medical treatment", "medicine"),
            ("Abstract query", "consciousness and existence", "philosophy")
        ]
        
        correct = 0
        for query_name, query_text, expected in queries:
            results = collection.search_text(query_text, limit=3)
            
            if results:
                top_result = results[0]
                is_correct = top_result.id == expected
                
                print(f"\n{query_name}: '{query_text}'")
                print(f"   Expected: {expected}")
                
                if is_correct:
                    correct += 1
                    print(f"   ✅ CORRECT: {top_result.id} ({top_result.score:.4f})")
                else:
                    print(f"   ❌ WRONG: {top_result.id} ({top_result.score:.4f})")
                
                for i, result in enumerate(results):
                    marker = "👑" if i == 0 else "  "
                    print(f"   {marker} {i+1}. {result.id}: {result.score:.4f}")
        
        accuracy = correct / len(queries)
        print(f"\n📊 OpenAI Accuracy: {accuracy:.1%} ({correct}/{len(queries)})")
        
        if accuracy >= 0.8:
            print("🎉 EXCELLENT - OpenAI provides high-quality embeddings!")
        elif accuracy >= 0.6:
            print("✅ GOOD - OpenAI works well")
        else:
            print("⚠️  FAIR - Results could be better")
            
    except Exception as e:
        print(f"❌ OpenAI test failed: {e}")

if __name__ == "__main__":
    test_openai_with_key()
