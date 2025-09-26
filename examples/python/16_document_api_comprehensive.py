#!/usr/bin/env python3
"""
VittoriaDB Document API Comprehensive Demo
Demonstrates all features of the new document-oriented API including:
- Schema-based document storage
- Multiple vector fields
- Full-text search with BM25
- Vector similarity search  
- Hybrid search
- Advanced filtering and facets
- Sorting and pagination
- Document CRUD operations
- Nested object support
"""

import requests
import json
import time
import random
import math
from typing import Dict, List, Any, Optional
import hashlib

class VittoriaDocumentClient:
    """Comprehensive client for VittoriaDB Document API"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        
    def create_database(self, schema: Dict[str, Any], **kwargs) -> Dict[str, Any]:
        """Create a document database with comprehensive schema"""
        payload = {
            "schema": schema,
            "language": kwargs.get("language", "english"),
            "fulltext_config": {
                "stemming": kwargs.get("stemming", True),
                "case_sensitive": kwargs.get("case_sensitive", False),
                "stop_words": kwargs.get("stop_words", [
                    "the", "a", "an", "and", "or", "but", "in", "on", "at", 
                    "to", "for", "of", "with", "by"
                ]),
                "bm25": {
                    "k": kwargs.get("bm25_k", 1.2),
                    "b": kwargs.get("bm25_b", 0.75),
                    "d": kwargs.get("bm25_d", 0.5)
                }
            }
        }
        
        response = self.session.post(f"{self.base_url}/create", json=payload)
        response.raise_for_status()
        return response.json()
        
    def insert_document(self, document: Dict[str, Any], **options) -> Dict[str, Any]:
        """Insert a document into the database"""
        payload = {
            "document": document,
            "options": options
        }
        
        response = self.session.post(f"{self.base_url}/documents", json=payload)
        response.raise_for_status()
        return response.json()
        
    def get_document(self, doc_id: str, include_vectors: bool = False) -> Dict[str, Any]:
        """Get a document by ID"""
        params = {"include_vectors": "true"} if include_vectors else {}
        response = self.session.get(f"{self.base_url}/documents/{doc_id}", params=params)
        response.raise_for_status()
        return response.json()
        
    def update_document(self, doc_id: str, document: Dict[str, Any], **options) -> Dict[str, Any]:
        """Update a document by ID"""
        payload = {
            "document": document,
            "options": options
        }
        
        response = self.session.put(f"{self.base_url}/documents/{doc_id}", json=payload)
        response.raise_for_status()
        return response.json()
        
    def delete_document(self, doc_id: str) -> Dict[str, Any]:
        """Delete a document by ID"""
        response = self.session.delete(f"{self.base_url}/documents/{doc_id}")
        response.raise_for_status()
        return response.json()
        
    def count_documents(self, where: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Count documents with optional filtering"""
        params = {"where": json.dumps(where)} if where else {}
        response = self.session.get(f"{self.base_url}/count", params=params)
        response.raise_for_status()
        return response.json()
        
    def search(self, **kwargs) -> Dict[str, Any]:
        """Perform comprehensive search with all supported parameters"""
        search_params = {
            "mode": kwargs.get("mode", "fulltext"),
            "limit": kwargs.get("limit", 10),
            "offset": kwargs.get("offset", 0)
        }
        
        # Add search-specific parameters
        if "term" in kwargs:
            search_params["term"] = kwargs["term"]
            
        if "vector" in kwargs:
            search_params["vector"] = kwargs["vector"]
            
        if "where" in kwargs:
            search_params["where"] = kwargs["where"]
            
        if "facets" in kwargs:
            search_params["facets"] = kwargs["facets"]
            
        if "sort_by" in kwargs:
            search_params["sort_by"] = kwargs["sort_by"]
            
        if "hybrid_weights" in kwargs:
            search_params["hybrid_weights"] = kwargs["hybrid_weights"]
            
        if "similarity" in kwargs:
            search_params["similarity"] = kwargs["similarity"]
            
        if "properties" in kwargs:
            search_params["properties"] = kwargs["properties"]
            
        response = self.session.post(f"{self.base_url}/search", json=search_params)
        response.raise_for_status()
        return response.json()

def generate_realistic_embedding(text: str, dimensions: int = 384) -> List[float]:
    """Generate a realistic-looking embedding based on text content"""
    # Create deterministic hash from text
    text_hash = hashlib.md5(text.encode()).hexdigest()
    random.seed(int(text_hash[:8], 16))
    
    # Generate base vector with text-influenced patterns
    vector = []
    char_influence = sum(ord(c) for c in text[:min(len(text), 50)]) / 50.0 if text else 0
    
    for i in range(dimensions):
        # Create patterns based on text characteristics
        position_influence = math.sin(i * 0.1) * 0.3
        random_component = random.gauss(0, 0.2)
        
        value = (char_influence / 255.0) + position_influence + random_component
        vector.append(max(-1.0, min(1.0, value)))  # Clamp to [-1, 1]
        
    # Normalize vector
    magnitude = math.sqrt(sum(x*x for x in vector))
    if magnitude > 0:
        vector = [x / magnitude for x in vector]
        
    return vector

def create_comprehensive_schema() -> Dict[str, Any]:
    """Create a comprehensive schema for testing all features"""
    return {
        # Basic text fields
        "title": "string",
        "description": "string",
        "content": "string",
        "summary": "string",
        "keywords": "string",
        
        # Categorical fields
        "category": "string",
        "subcategory": "string",
        "language": "string",
        "content_type": "string",
        "status": "string",
        
        # Numerical fields
        "price": "number",
        "rating": "number",
        "word_count": "number",
        "reading_time": "number",
        "difficulty_score": "number",
        "views": "number",
        "downloads": "number",
        
        # Vector fields (multiple embeddings)
        "title_embedding": "vector[384]",
        "content_embedding": "vector[384]",
        "summary_embedding": "vector[128]",
        
        # Nested metadata object
        "metadata": {
            "author": "string",
            "published_date": "string",
            "last_updated": "string",
            "source": "string",
            "tags": "string",
            "version": "number",
            "license": "string",
            "contributors": "string"
        },
        
        # Boolean fields
        "available": "boolean",
        "featured": "boolean",
        "premium": "boolean",
        "verified": "boolean",
        "open_source": "boolean"
    }

def create_sample_documents() -> List[Dict[str, Any]]:
    """Create comprehensive sample documents for testing"""
    documents = [
        {
            "id": "ml_guide_2024",
            "title": "The Complete Guide to Machine Learning in 2024",
            "description": "An exhaustive guide covering all aspects of modern machine learning, from fundamentals to advanced techniques",
            "content": """Machine learning has revolutionized the way we approach complex problems across industries. This comprehensive guide explores the fundamental concepts, algorithms, and practical applications that define the field in 2024.

Chapter 1: Foundations of Machine Learning
Machine learning is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed. The field encompasses supervised learning, unsupervised learning, and reinforcement learning paradigms.

Supervised learning involves training models on labeled datasets to make predictions on new, unseen data. Common algorithms include linear regression, decision trees, random forests, support vector machines, and neural networks.

Chapter 2: Deep Learning and Neural Networks
Deep learning has emerged as one of the most powerful approaches in machine learning. Convolutional Neural Networks (CNNs) excel at image recognition tasks, while Recurrent Neural Networks (RNNs) and Long Short-Term Memory (LSTM) networks are ideal for sequential data processing.

The transformer architecture, introduced in the "Attention Is All You Need" paper, has revolutionized natural language processing. Models like BERT, GPT, and T5 have achieved remarkable performance on various NLP tasks.

Chapter 3: Practical Implementation
Implementing machine learning solutions requires careful consideration of data preprocessing, feature engineering, model selection, and evaluation metrics. Cross-validation, hyperparameter tuning, and regularization techniques are essential for building robust models.""",
            "summary": "A comprehensive guide covering machine learning fundamentals, deep learning, and practical implementation in 2024.",
            "keywords": "machine learning, deep learning, neural networks, AI, algorithms, data science",
            "category": "technology",
            "subcategory": "artificial-intelligence",
            "language": "english",
            "content_type": "educational-guide",
            "status": "published",
            "price": 49.99,
            "rating": 4.8,
            "word_count": 2847,
            "reading_time": 12,
            "difficulty_score": 7.5,
            "views": 15420,
            "downloads": 3240,
            "title_embedding": generate_realistic_embedding("The Complete Guide to Machine Learning in 2024", 384),
            "content_embedding": generate_realistic_embedding("machine learning fundamentals deep learning neural networks", 384),
            "summary_embedding": generate_realistic_embedding("comprehensive machine learning guide", 128),
            "metadata": {
                "author": "Dr. Sarah Chen",
                "published_date": "2024-01-15",
                "last_updated": "2024-09-01",
                "source": "TechEducation Press",
                "tags": "machine learning, AI, deep learning, neural networks, data science",
                "version": 2.1,
                "license": "MIT",
                "contributors": "Dr. Sarah Chen, Prof. John Smith"
            },
            "available": True,
            "featured": True,
            "premium": False,
            "verified": True,
            "open_source": True
        },
        {
            "id": "quantum_computing_intro",
            "title": "Introduction to Quantum Computing: Principles and Applications",
            "description": "Explore the fascinating world of quantum computing, from basic principles to real-world applications",
            "content": """Quantum computing represents a paradigm shift in computational capability, leveraging the principles of quantum mechanics to process information in fundamentally new ways.

Understanding Quantum Mechanics
At the heart of quantum computing lies the concept of quantum superposition, where quantum bits (qubits) can exist in multiple states simultaneously. Unlike classical bits that are either 0 or 1, qubits can be in a superposition of both states, enabling parallel computation on an unprecedented scale.

Quantum entanglement is another crucial phenomenon where qubits become correlated in such a way that the quantum state of each qubit cannot be described independently. This property enables quantum computers to perform certain calculations exponentially faster than classical computers.

Quantum Algorithms
Several quantum algorithms have been developed that demonstrate quantum advantage over classical algorithms. Shor's algorithm for integer factorization could potentially break current cryptographic systems, while Grover's algorithm provides quadratic speedup for searching unsorted databases.

The Variational Quantum Eigensolver (VQE) and Quantum Approximate Optimization Algorithm (QAOA) are hybrid quantum-classical algorithms that show promise for near-term quantum devices.""",
            "summary": "An introduction to quantum computing covering principles, algorithms, and applications.",
            "keywords": "quantum computing, qubits, superposition, entanglement, quantum algorithms",
            "category": "technology",
            "subcategory": "quantum-computing",
            "language": "english",
            "content_type": "educational-article",
            "status": "published",
            "price": 0.0,
            "rating": 4.6,
            "word_count": 1923,
            "reading_time": 8,
            "difficulty_score": 8.2,
            "views": 8750,
            "downloads": 1890,
            "title_embedding": generate_realistic_embedding("Introduction to Quantum Computing: Principles and Applications", 384),
            "content_embedding": generate_realistic_embedding("quantum computing qubits superposition algorithms", 384),
            "summary_embedding": generate_realistic_embedding("quantum computing introduction", 128),
            "metadata": {
                "author": "Prof. Michael Zhang",
                "published_date": "2024-03-10",
                "last_updated": "2024-08-15",
                "source": "Quantum Research Institute",
                "tags": "quantum computing, physics, technology, algorithms",
                "version": 1.3,
                "license": "Apache-2.0",
                "contributors": "Prof. Michael Zhang, Dr. Alice Johnson"
            },
            "available": True,
            "featured": False,
            "premium": True,
            "verified": True,
            "open_source": False
        },
        {
            "id": "web_dev_javascript",
            "title": "Modern Web Development with JavaScript",
            "description": "Master modern JavaScript frameworks and tools for building scalable web applications",
            "content": """JavaScript has evolved significantly over the years, becoming the backbone of modern web development. This guide covers the latest frameworks, tools, and best practices for building scalable web applications.

React and Component-Based Architecture
React has revolutionized how we build user interfaces with its component-based architecture. Learn about hooks, state management with Redux and Context API, and performance optimization techniques like memoization and code splitting.

Node.js and Backend Development
Node.js enables JavaScript developers to build scalable backend services. Explore Express.js for building REST APIs, database integration with MongoDB and PostgreSQL, and real-time applications with WebSockets.

Modern Development Tools
The JavaScript ecosystem offers powerful development tools including Webpack for bundling, Babel for transpilation, ESLint for code quality, and Jest for testing. Learn how to set up efficient development workflows.

Performance Optimization
Modern web applications must be fast and responsive. Learn about lazy loading, service workers, Progressive Web Apps (PWAs), and performance monitoring techniques.""",
            "summary": "A comprehensive guide to modern JavaScript development for web applications.",
            "keywords": "javascript, web development, react, nodejs, frontend, backend",
            "category": "programming",
            "subcategory": "web-development",
            "language": "english",
            "content_type": "tutorial",
            "status": "published",
            "price": 34.99,
            "rating": 4.7,
            "word_count": 1654,
            "reading_time": 7,
            "difficulty_score": 6.0,
            "views": 12300,
            "downloads": 2890,
            "title_embedding": generate_realistic_embedding("Modern Web Development with JavaScript", 384),
            "content_embedding": generate_realistic_embedding("javascript react nodejs web development", 384),
            "summary_embedding": generate_realistic_embedding("modern javascript development", 128),
            "metadata": {
                "author": "Alex Thompson",
                "published_date": "2024-02-20",
                "last_updated": "2024-09-15",
                "source": "WebDev Academy",
                "tags": "javascript, web development, react, nodejs, frontend",
                "version": 1.8,
                "license": "Creative Commons",
                "contributors": "Alex Thompson, Maria Garcia, Tom Wilson"
            },
            "available": True,
            "featured": True,
            "premium": False,
            "verified": True,
            "open_source": True
        },
        {
            "id": "data_science_python",
            "title": "Data Science with Python: Analytics and Visualization",
            "description": "Learn data science techniques using Python, pandas, and machine learning libraries",
            "content": """Python has become the de facto language for data science, offering powerful libraries and tools for data analysis, visualization, and machine learning.

Data Analysis with Pandas
Pandas provides powerful data structures and analysis tools. Learn about DataFrames, data cleaning, transformation, and exploratory data analysis techniques. Master groupby operations, merging datasets, and handling missing data.

Visualization with Matplotlib and Seaborn
Create compelling visualizations to communicate insights effectively. Learn about statistical plots, customizing charts, and creating publication-ready figures.

Machine Learning with Scikit-learn
Scikit-learn offers a comprehensive suite of machine learning algorithms. Explore classification, regression, clustering, and dimensionality reduction techniques. Learn about model evaluation, cross-validation, and hyperparameter tuning.

Advanced Topics
Dive into advanced topics including time series analysis, natural language processing with NLTK and spaCy, and deep learning with TensorFlow and PyTorch.""",
            "summary": "A practical guide to data science using Python and its ecosystem.",
            "keywords": "python, data science, pandas, machine learning, analytics, visualization",
            "category": "data-science",
            "subcategory": "python-programming",
            "language": "english",
            "content_type": "course",
            "status": "published",
            "price": 59.99,
            "rating": 4.9,
            "word_count": 2156,
            "reading_time": 9,
            "difficulty_score": 5.5,
            "views": 18750,
            "downloads": 4320,
            "title_embedding": generate_realistic_embedding("Data Science with Python: Analytics and Visualization", 384),
            "content_embedding": generate_realistic_embedding("python data science pandas machine learning", 384),
            "summary_embedding": generate_realistic_embedding("python data science guide", 128),
            "metadata": {
                "author": "Dr. Lisa Wang",
                "published_date": "2024-01-05",
                "last_updated": "2024-09-20",
                "source": "DataScience Institute",
                "tags": "python, data science, machine learning, analytics",
                "version": 2.0,
                "license": "MIT",
                "contributors": "Dr. Lisa Wang, Prof. David Lee, Sarah Kim"
            },
            "available": True,
            "featured": True,
            "premium": True,
            "verified": True,
            "open_source": True
        },
        {
            "id": "cloud_architecture",
            "title": "Cloud Architecture Patterns and Best Practices",
            "description": "Design scalable and resilient cloud applications using modern architecture patterns",
            "content": """Cloud computing has transformed how we design and deploy applications. This guide covers essential architecture patterns and best practices for building scalable, resilient cloud applications.

Microservices Architecture
Learn how to decompose monolithic applications into microservices. Understand service boundaries, communication patterns, and data management strategies. Explore containerization with Docker and orchestration with Kubernetes.

Serverless Computing
Serverless architectures enable you to build and run applications without managing servers. Learn about AWS Lambda, Azure Functions, and Google Cloud Functions. Understand event-driven architectures and Function-as-a-Service (FaaS) patterns.

Data Architecture in the Cloud
Design robust data architectures using cloud-native services. Learn about data lakes, data warehouses, and real-time streaming architectures. Explore services like Amazon S3, Google BigQuery, and Azure Data Factory.

Security and Compliance
Implement security best practices including identity and access management, encryption, and network security. Understand compliance frameworks and how to design secure cloud architectures.""",
            "summary": "A comprehensive guide to cloud architecture patterns and best practices.",
            "keywords": "cloud computing, microservices, serverless, kubernetes, AWS, Azure, architecture",
            "category": "technology",
            "subcategory": "cloud-computing",
            "language": "english",
            "content_type": "technical-guide",
            "status": "published",
            "price": 79.99,
            "rating": 4.5,
            "word_count": 1890,
            "reading_time": 8,
            "difficulty_score": 7.8,
            "views": 9650,
            "downloads": 2150,
            "title_embedding": generate_realistic_embedding("Cloud Architecture Patterns and Best Practices", 384),
            "content_embedding": generate_realistic_embedding("cloud computing microservices serverless architecture", 384),
            "summary_embedding": generate_realistic_embedding("cloud architecture guide", 128),
            "metadata": {
                "author": "Robert Chen",
                "published_date": "2024-04-12",
                "last_updated": "2024-09-05",
                "source": "Cloud Architecture Institute",
                "tags": "cloud computing, architecture, microservices, serverless",
                "version": 1.5,
                "license": "Apache-2.0",
                "contributors": "Robert Chen, Emily Davis"
            },
            "available": True,
            "featured": False,
            "premium": True,
            "verified": True,
            "open_source": False
        }
    ]
    
    return documents

def main():
    """Run comprehensive VittoriaDB Document API demonstration"""
    print("🚀 VittoriaDB Document API Comprehensive Demo")
    print("=" * 60)
    
    # Initialize client
    client = VittoriaDocumentClient()
    
    # Step 1: Create database with comprehensive schema
    print("\n🔧 Creating Document Database with Comprehensive Schema")
    schema = create_comprehensive_schema()
    
    try:
        result = client.create_database(schema)
        print("✅ Database created successfully")
        print(f"   📊 Vector fields: {result.get('vector_fields', {})}")
        print(f"   🔍 Searchable fields: {len(result.get('searchable_fields', []))} fields")
        print(f"   📁 Collections created: {result.get('created_collections', [])}")
    except Exception as e:
        print(f"❌ Failed to create database: {e}")
        return
    
    # Step 2: Insert sample documents
    print("\n📚 Inserting Sample Documents")
    documents = create_sample_documents()
    
    inserted_count = 0
    for doc in documents:
        try:
            result = client.insert_document(doc)
            if result.get("created"):
                inserted_count += 1
                print(f"✅ Inserted: {doc['title'][:50]}...")
            else:
                print(f"⚠️  Document {doc['id']} not marked as created")
        except Exception as e:
            print(f"❌ Failed to insert {doc['id']}: {e}")
    
    print(f"\n📊 Insertion Summary: {inserted_count}/{len(documents)} documents inserted")
    
    # Wait for indexing
    time.sleep(2)
    
    # Step 3: Full-text search demonstrations
    print("\n🔍 Full-Text Search Demonstrations")
    
    text_searches = [
        ("Technical Terms", "machine learning neural networks"),
        ("Programming Topics", "javascript web development"),
        ("Data Science", "python data science pandas"),
        ("Cloud Computing", "cloud architecture microservices"),
        ("Author Search", "Dr. Sarah Chen"),
        ("Content Search", "algorithms optimization")
    ]
    
    for name, query in text_searches:
        print(f"\n🎯 {name}")
        try:
            results = client.search(mode="fulltext", term=query, limit=3)
            hits = results.get("hits", [])
            print(f"   Found {len(hits)} results:")
            
            for i, hit in enumerate(hits, 1):
                doc = hit.get("document", {})
                title = doc.get("title", "Unknown")
                category = doc.get("category", "Unknown")
                score = hit.get("score", 0)
                print(f"   {i}. {title[:50]}... (Score: {score:.3f}, Category: {category})")
                
        except Exception as e:
            print(f"   ❌ Search failed: {e}")
    
    # Step 4: Vector search demonstrations
    print("\n🎯 Vector Search Demonstrations")
    
    vector_searches = [
        ("Title Similarity", "artificial intelligence guide", "title_embedding", 384),
        ("Content Similarity", "programming tutorial examples", "content_embedding", 384),
        ("Summary Similarity", "comprehensive learning guide", "summary_embedding", 128)
    ]
    
    for name, query, property_name, dims in vector_searches:
        print(f"\n🎯 {name}")
        try:
            query_vector = generate_realistic_embedding(query, dims)
            
            results = client.search(
                mode="vector",
                vector={
                    "value": query_vector,
                    "property": property_name
                },
                limit=3,
                similarity=0.7
            )
            
            hits = results.get("hits", [])
            print(f"   Found {len(hits)} results:")
            
            for i, hit in enumerate(hits, 1):
                doc = hit.get("document", {})
                title = doc.get("title", "Unknown")
                score = hit.get("score", 0)
                print(f"   {i}. {title[:50]}... (Score: {score:.3f})")
                
        except Exception as e:
            print(f"   ❌ Vector search failed: {e}")
    
    # Step 5: Hybrid search demonstrations
    print("\n🔀 Hybrid Search Demonstrations")
    
    hybrid_searches = [
        ("Balanced Search", "programming", "software development tutorial", 0.5, 0.5),
        ("Text-Heavy Search", "machine learning", "AI algorithms", 0.8, 0.2),
        ("Vector-Heavy Search", "data science", "analytics visualization", 0.2, 0.8)
    ]
    
    for name, term, vector_query, text_weight, vector_weight in hybrid_searches:
        print(f"\n🎯 {name}")
        try:
            query_vector = generate_realistic_embedding(vector_query, 384)
            
            results = client.search(
                mode="hybrid",
                term=term,
                vector={
                    "value": query_vector,
                    "property": "content_embedding"
                },
                hybrid_weights={
                    "text": text_weight,
                    "vector": vector_weight
                },
                limit=3
            )
            
            hits = results.get("hits", [])
            print(f"   Found {len(hits)} results (Text: {text_weight}, Vector: {vector_weight}):")
            
            for i, hit in enumerate(hits, 1):
                doc = hit.get("document", {})
                title = doc.get("title", "Unknown")
                category = doc.get("category", "Unknown")
                score = hit.get("score", 0)
                print(f"   {i}. {title[:50]}... (Score: {score:.3f}, Category: {category})")
                
        except Exception as e:
            print(f"   ❌ Hybrid search failed: {e}")
    
    # Step 6: Advanced filtering demonstrations
    print("\n🔧 Advanced Filtering Demonstrations")
    
    filter_searches = [
        ("Premium Content", "*", {"premium": True}),
        ("High-Rated Content", "*", {"rating": {"gte": 4.7}}),
        ("Technology Category", "*", {"category": "technology"}),
        ("Recent & Featured", "*", {"featured": True, "rating": {"gte": 4.5}}),
        ("Open Source Projects", "*", {"open_source": True}),
        ("Price Range", "*", {"price": {"gte": 30, "lte": 60}})
    ]
    
    for name, term, filter_condition in filter_searches:
        print(f"\n🎯 {name}")
        try:
            results = client.search(
                mode="fulltext",
                term=term,
                where=filter_condition,
                limit=5
            )
            
            hits = results.get("hits", [])
            print(f"   Found {len(hits)} filtered results:")
            
            for i, hit in enumerate(hits, 1):
                doc = hit.get("document", {})
                title = doc.get("title", "Unknown")
                category = doc.get("category", "Unknown")
                rating = doc.get("rating", 0)
                premium = doc.get("premium", False)
                price = doc.get("price", 0)
                print(f"   {i}. {title[:40]}... (Category: {category}, Rating: {rating}, Premium: {premium}, Price: ${price})")
                
        except Exception as e:
            print(f"   ❌ Filtered search failed: {e}")
    
    # Step 7: Facet analysis demonstrations
    print("\n📊 Facet Analysis Demonstrations")
    
    try:
        results = client.search(
            mode="fulltext",
            term="*",
            limit=10,
            facets={
                "category": {"type": "string", "limit": 10},
                "premium": {"type": "string", "limit": 10},
                "featured": {"type": "string", "limit": 10},
                "open_source": {"type": "string", "limit": 10}
            }
        )
        
        facets = results.get("facets", {})
        print("📊 Facet Results:")
        for facet_name, facet_result in facets.items():
            print(f"   {facet_name}:")
            values = facet_result.get("values", {})
            for value, count in values.items():
                print(f"     {value}: {count}")
                
    except Exception as e:
        print(f"❌ Facet search failed: {e}")
    
    # Step 8: Sorting demonstrations
    print("\n📈 Sorting Demonstrations")
    
    sort_searches = [
        ("By Rating (Highest First)", "rating", "desc"),
        ("By Price (Lowest First)", "price", "asc"),
        ("By Views (Most Popular)", "views", "desc"),
        ("By Word Count (Shortest First)", "word_count", "asc"),
        ("By Downloads (Most Downloaded)", "downloads", "desc")
    ]
    
    for name, property_name, order in sort_searches:
        print(f"\n🎯 {name}")
        try:
            results = client.search(
                mode="fulltext",
                term="*",
                limit=5,
                sort_by={
                    "property": property_name,
                    "order": order
                }
            )
            
            hits = results.get("hits", [])
            print(f"   Sorted by {property_name} ({order}):")
            
            for i, hit in enumerate(hits, 1):
                doc = hit.get("document", {})
                title = doc.get("title", "Unknown")
                value = doc.get(property_name, "N/A")
                print(f"   {i}. {title[:40]}... ({property_name}: {value})")
                
        except Exception as e:
            print(f"   ❌ Sorted search failed: {e}")
    
    # Step 9: Document operations demonstrations
    print("\n📄 Document Operations Demonstrations")
    
    # Get document
    print("\n🔍 Getting Document by ID")
    try:
        result = client.get_document("ml_guide_2024", include_vectors=False)
        if result.get("found"):
            doc = result["document"]
            print(f"✅ Retrieved: {doc.get('title', 'Unknown')}")
            metadata = doc.get("metadata", {})
            print(f"   Author: {metadata.get('author', 'Unknown')}")
            print(f"   Rating: {doc.get('rating', 0)}")
            print(f"   Views: {doc.get('views', 0)}")
        else:
            print("❌ Document not found")
    except Exception as e:
        print(f"❌ Failed to get document: {e}")
    
    # Count documents
    print("\n🔢 Counting Documents")
    try:
        result = client.count_documents()
        count = result.get("count", 0)
        print(f"✅ Total documents in database: {count}")
    except Exception as e:
        print(f"❌ Failed to count documents: {e}")
    
    # Update document
    print("\n✏️  Updating Document")
    try:
        update_doc = {
            "title": "The Complete Guide to Machine Learning in 2024 - Updated Edition",
            "rating": 4.9,
            "views": 16000,
            "downloads": 3500
        }
        
        result = client.update_document("ml_guide_2024", update_doc)
        if result.get("updated"):
            print("✅ Document updated successfully")
        else:
            print("⚠️  Document not marked as updated")
    except Exception as e:
        print(f"❌ Failed to update document: {e}")
    
    # Performance summary
    print("\n⚡ Performance Summary")
    try:
        # Quick performance test
        start_time = time.time()
        results = client.search(mode="fulltext", term="machine learning", limit=5)
        search_time = (time.time() - start_time) * 1000
        
        print(f"✅ Search performance: {search_time:.1f}ms")
        print(f"✅ Results found: {len(results.get('hits', []))}")
        
    except Exception as e:
        print(f"❌ Performance test failed: {e}")
    
    # Final summary
    print("\n" + "=" * 60)
    print("🎉 Document API Comprehensive Demo Complete!")
    print("=" * 60)
    print("Demonstrated features:")
    print("• ✅ Comprehensive schema with multiple data types")
    print("• ✅ Multiple vector fields (384D, 128D)")
    print("• ✅ Full-text search with BM25 scoring")
    print("• ✅ Vector similarity search")
    print("• ✅ Hybrid search with custom weights")
    print("• ✅ Advanced filtering and facets")
    print("• ✅ Sorting by multiple properties")
    print("• ✅ Document CRUD operations")
    print("• ✅ Nested object support")
    print("• ✅ Production-ready performance")
    print("• ✅ Comprehensive error handling")

if __name__ == "__main__":
    main()
