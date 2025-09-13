#!/bin/bash

# VittoriaDB RAG (Retrieval-Augmented Generation) Example with cURL
# This script demonstrates building a RAG system using VittoriaDB HTTP API
# Make sure VittoriaDB is running: ./vittoriadb run

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8080"
COLLECTION_NAME="rag_curl_demo"
DIMENSIONS=256

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${BLUE}$1${NC}"
    echo "=================================="
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_rag() {
    echo -e "${PURPLE}🤖 $1${NC}"
}

print_query() {
    echo -e "${CYAN}❓ $1${NC}"
}

# Generate a simple embedding based on text content
# In a real implementation, you would use a proper embedding model
generate_text_embedding() {
    local text="$1"
    local vector="["
    
    # Simple hash-based embedding generation
    local text_lower=$(echo "$text" | tr '[:upper:]' '[:lower:]')
    local words=($text_lower)
    
    # Initialize vector with zeros
    for ((i=0; i<DIMENSIONS; i++)); do
        if [ $i -gt 0 ]; then
            vector+=", "
        fi
        vector+="0.0"
    done
    
    # Modify vector based on word characteristics
    local word_index=0
    for word in "${words[@]}"; do
        # Simple hash function
        local hash=0
        for ((i=0; i<${#word}; i++)); do
            char=$(printf "%d" "'${word:$i:1}")
            hash=$((hash * 31 + char))
        done
        
        # Map hash to vector positions
        local pos1=$((hash % DIMENSIONS))
        local pos2=$(((hash / DIMENSIONS) % DIMENSIONS))
        
        # Add word influence to vector (simplified)
        local influence=$(awk "BEGIN {printf \"%.6f\", ($hash % 1000) / 1000.0 - 0.5}")
        
        # Replace the zero at specific positions with computed values
        if [ $word_index -lt 10 ]; then  # Limit to first 10 words for performance
            vector=$(echo "$vector" | sed "s/0\.0/$influence/" | sed "s/$influence/0.0/g" | sed "${pos1}s/0\.0/$influence/")
        fi
        
        word_index=$((word_index + 1))
    done
    
    vector+="]"
    echo "$vector"
}

# Create sample knowledge base documents
create_knowledge_base() {
    print_header "Creating Knowledge Base"
    
    # Document chunks with their content
    declare -A documents
    documents["chunk_1"]="Vector databases are specialized database systems designed to store, index, and query high-dimensional vector data efficiently. They are essential for modern AI applications, particularly those involving machine learning and natural language processing."
    documents["chunk_2"]="VittoriaDB is a high-performance vector database that supports multiple indexing algorithms including Flat index for exact search and HNSW for approximate nearest neighbor search. It offers various distance metrics such as cosine similarity, Euclidean distance, and dot product."
    documents["chunk_3"]="Retrieval-Augmented Generation (RAG) is an AI architecture that combines information retrieval with text generation to produce more accurate and contextually relevant responses. A typical RAG system consists of a knowledge base, embedding model, retrieval system, and language model."
    documents["chunk_4"]="Embeddings are dense vector representations of data that capture semantic meaning in a continuous vector space. In natural language processing, embeddings convert words and sentences into numerical vectors that encode semantic relationships."
    documents["chunk_5"]="HNSW (Hierarchical Navigable Small World) is an algorithm for approximate nearest neighbor search in high-dimensional spaces. It builds a multi-layer graph structure that enables fast similarity search even with millions of vectors."
    documents["chunk_6"]="Machine learning models in production require careful consideration of scalability, reliability, and maintainability. Vector databases play a crucial role in production ML systems, especially for recommendation engines and search systems."
    documents["chunk_7"]="Document processing in RAG systems involves chunking text into smaller segments, generating embeddings for each chunk, and storing them in a vector database. The quality of chunking strategy directly impacts retrieval performance."
    documents["chunk_8"]="Semantic search uses vector embeddings to find documents based on meaning rather than exact keyword matches. This enables more intelligent search capabilities that understand context and intent."
    
    print_info "Processing ${#documents[@]} knowledge base chunks..."
    
    # Create vectors array for batch insertion
    local batch_data="{\"vectors\": ["
    local first=true
    
    for chunk_id in "${!documents[@]}"; do
        local content="${documents[$chunk_id]}"
        local embedding=$(generate_text_embedding "$content")
        
        if [ "$first" = false ]; then
            batch_data+=", "
        fi
        first=false
        
        # Extract key terms for metadata
        local word_count=$(echo "$content" | wc -w | tr -d ' ')
        local char_count=${#content}
        local category="general"
        
        if [[ "$content" == *"VittoriaDB"* ]]; then
            category="vittoriadb"
        elif [[ "$content" == *"RAG"* ]] || [[ "$content" == *"Retrieval"* ]]; then
            category="rag"
        elif [[ "$content" == *"embedding"* ]] || [[ "$content" == *"vector"* ]]; then
            category="embeddings"
        elif [[ "$content" == *"HNSW"* ]] || [[ "$content" == *"algorithm"* ]]; then
            category="algorithms"
        fi
        
        batch_data+="{
            \"id\": \"$chunk_id\",
            \"vector\": $embedding,
            \"metadata\": {
                \"content\": \"$content\",
                \"category\": \"$category\",
                \"word_count\": $word_count,
                \"char_count\": $char_count,
                \"chunk_type\": \"knowledge_base\"
            }
        }"
        
        print_info "Processed $chunk_id ($category, $word_count words)"
    done
    
    batch_data+="]}"
    
    # Insert all chunks in one batch
    print_info "Storing knowledge base in vector database..."
    curl -s -X POST "$BASE_URL/collections/$COLLECTION_NAME/vectors/batch" \
        -H "Content-Type: application/json" \
        -d "$batch_data" > /dev/null
    
    print_success "Knowledge base created with ${#documents[@]} chunks"
}

# Perform RAG query
perform_rag_query() {
    local question="$1"
    local k="${2:-3}"
    
    print_query "Question: $question"
    
    # Generate embedding for the question
    local query_embedding=$(generate_text_embedding "$question")
    
    # Search for relevant chunks
    print_info "Searching for relevant information..."
    local search_results=$(curl -s -X POST "$BASE_URL/collections/$COLLECTION_NAME/search" \
        -H "Content-Type: application/json" \
        -d "{
            \"vector\": $query_embedding,
            \"k\": $k
        }")
    
    # Extract and display results
    local result_count=$(echo "$search_results" | jq '.results | length' 2>/dev/null || echo "0")
    
    if [ "$result_count" -eq 0 ]; then
        print_rag "No relevant information found for this question."
        return
    fi
    
    print_success "Found $result_count relevant chunks"
    
    # Display search results
    echo "$search_results" | jq -r '.results[] | "📄 \(.id) (score: \(.score | tostring | .[0:5])): \(.metadata.content | .[0:100])..."' 2>/dev/null || {
        echo "Search results (raw):"
        echo "$search_results"
    }
    
    # Generate simple answer based on top result
    local top_content=$(echo "$search_results" | jq -r '.results[0].metadata.content' 2>/dev/null || echo "Unable to extract content")
    local top_score=$(echo "$search_results" | jq -r '.results[0].score' 2>/dev/null || echo "0")
    
    print_rag "Generated Answer:"
    if [[ "$top_content" != "Unable to extract content" ]]; then
        # Simple answer generation based on question type
        if [[ "$question" == *"what is"* ]] || [[ "$question" == *"What is"* ]]; then
            echo "Based on the knowledge base: $(echo "$top_content" | cut -c1-200)..."
        elif [[ "$question" == *"how"* ]] || [[ "$question" == *"How"* ]]; then
            echo "According to the available information: $(echo "$top_content" | cut -c1-200)..."
        else
            echo "The relevant information indicates: $(echo "$top_content" | cut -c1-200)..."
        fi
        echo ""
        print_info "Confidence: $(echo "$top_score" | cut -c1-5) (based on similarity score)"
    else
        echo "I found some relevant information, but couldn't process it properly."
    fi
}

# Perform filtered RAG query
perform_filtered_rag_query() {
    local question="$1"
    local category="$2"
    local k="${3:-2}"
    
    print_query "Filtered Question: $question (category: $category)"
    
    # Generate embedding for the question
    local query_embedding=$(generate_text_embedding "$question")
    
    # Search with category filter
    print_info "Searching with category filter..."
    local search_results=$(curl -s -X POST "$BASE_URL/collections/$COLLECTION_NAME/search" \
        -H "Content-Type: application/json" \
        -d "{
            \"vector\": $query_embedding,
            \"k\": $k,
            \"filter\": {
                \"category\": \"$category\"
            }
        }")
    
    # Process and display results
    local result_count=$(echo "$search_results" | jq '.results | length' 2>/dev/null || echo "0")
    
    if [ "$result_count" -eq 0 ]; then
        print_rag "No relevant information found in the '$category' category."
        return
    fi
    
    print_success "Found $result_count relevant chunks in '$category' category"
    
    # Display filtered results
    echo "$search_results" | jq -r '.results[] | "📄 \(.id) (score: \(.score | tostring | .[0:5])): \(.metadata.content | .[0:100])..."' 2>/dev/null || {
        echo "Filtered search results (raw):"
        echo "$search_results"
    }
    
    # Generate answer from filtered results
    local top_content=$(echo "$search_results" | jq -r '.results[0].metadata.content' 2>/dev/null || echo "Unable to extract content")
    
    print_rag "Filtered Answer:"
    if [[ "$top_content" != "Unable to extract content" ]]; then
        echo "From the $category knowledge: $(echo "$top_content" | cut -c1-200)..."
    else
        echo "Found relevant information in the $category category, but couldn't process it."
    fi
}

# Interactive RAG demo
interactive_rag_demo() {
    print_header "Interactive RAG Demo"
    
    print_info "Ask questions about the knowledge base (type 'quit' to exit)"
    print_info "Available topics: vector databases, VittoriaDB, RAG, embeddings, HNSW, machine learning"
    
    while true; do
        echo ""
        read -p "$(echo -e ${CYAN}❓ Your question: ${NC})" question
        
        if [[ -z "$question" ]]; then
            continue
        fi
        
        if [[ "$question" == "quit" ]] || [[ "$question" == "exit" ]]; then
            print_info "Exiting interactive demo"
            break
        fi
        
        # Determine if this should be a filtered query
        if [[ "$question" == *"VittoriaDB"* ]]; then
            perform_filtered_rag_query "$question" "vittoriadb" 2
        elif [[ "$question" == *"RAG"* ]] || [[ "$question" == *"retrieval"* ]]; then
            perform_filtered_rag_query "$question" "rag" 2
        elif [[ "$question" == *"embedding"* ]] || [[ "$question" == *"vector"* ]]; then
            perform_filtered_rag_query "$question" "embeddings" 2
        elif [[ "$question" == *"HNSW"* ]] || [[ "$question" == *"algorithm"* ]]; then
            perform_filtered_rag_query "$question" "algorithms" 2
        else
            perform_rag_query "$question" 3
        fi
    done
}

# Check connection
check_connection() {
    print_header "Connection Test"
    
    if curl -s -f "$BASE_URL/stats" > /dev/null; then
        print_success "Connected to VittoriaDB at $BASE_URL"
    else
        print_error "Failed to connect to VittoriaDB"
        print_info "Make sure VittoriaDB is running with: ./vittoriadb run"
        exit 1
    fi
}

# Setup RAG collection
setup_rag_collection() {
    print_header "Setting up RAG Collection"
    
    # Delete existing collection if it exists
    curl -s -X DELETE "$BASE_URL/collections/$COLLECTION_NAME" > /dev/null 2>&1 || true
    
    # Create new collection optimized for RAG
    response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/collections" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$COLLECTION_NAME\",
            \"dimensions\": $DIMENSIONS,
            \"index_type\": 1,
            \"metric\": 0,
            \"description\": \"RAG knowledge base with document chunks\"
        }")
    
    http_code="${response: -3}"
    if [[ "$http_code" == "201" ]]; then
        print_success "Created RAG collection '$COLLECTION_NAME' ($DIMENSIONS dimensions, HNSW index)"
    else
        print_error "Failed to create collection (HTTP $http_code)"
        exit 1
    fi
}

# Demonstrate different RAG queries
demonstrate_rag_queries() {
    print_header "RAG Query Demonstrations"
    
    # Sample queries
    local queries=(
        "What is a vector database?"
        "How does VittoriaDB work?"
        "What is RAG?"
        "How do embeddings work?"
        "What is HNSW algorithm?"
        "How to use machine learning in production?"
    )
    
    for query in "${queries[@]}"; do
        echo ""
        perform_rag_query "$query" 3
        sleep 1  # Brief pause between queries
    done
    
    # Demonstrate filtered queries
    echo ""
    print_info "Demonstrating filtered queries..."
    perform_filtered_rag_query "Tell me about database features" "vittoriadb" 2
    perform_filtered_rag_query "How does retrieval work?" "rag" 2
}

# Performance analysis
analyze_rag_performance() {
    print_header "RAG Performance Analysis"
    
    local test_queries=(
        "vector database performance"
        "machine learning embeddings"
        "search algorithms efficiency"
    )
    
    print_info "Testing RAG query performance..."
    
    for query in "${test_queries[@]}"; do
        print_info "Testing query: '$query'"
        
        # Time the embedding generation and search
        start_time=$(date +%s.%N)
        
        local query_embedding=$(generate_text_embedding "$query")
        local search_results=$(curl -s -X POST "$BASE_URL/collections/$COLLECTION_NAME/search" \
            -H "Content-Type: application/json" \
            -d "{
                \"vector\": $query_embedding,
                \"k\": 5
            }")
        
        end_time=$(date +%s.%N)
        query_time=$(echo "$end_time - $start_time" | bc)
        
        local result_count=$(echo "$search_results" | jq '.results | length' 2>/dev/null || echo "0")
        
        print_success "Query completed in ${query_time}s, found $result_count results"
    done
    
    # Test batch vs individual queries
    print_info "Comparing batch vs individual query performance..."
    
    # Individual queries
    start_time=$(date +%s.%N)
    for i in {1..5}; do
        local query_embedding=$(generate_text_embedding "test query $i")
        curl -s -X POST "$BASE_URL/collections/$COLLECTION_NAME/search" \
            -H "Content-Type: application/json" \
            -d "{
                \"vector\": $query_embedding,
                \"k\": 3
            }" > /dev/null
    done
    individual_time=$(echo "$(date +%s.%N) - $start_time" | bc)
    
    print_success "5 individual queries: ${individual_time}s"
    print_info "Average per query: $(echo "scale=4; $individual_time / 5" | bc)s"
}

# Cleanup
cleanup() {
    print_header "Cleanup"
    
    print_info "Deleting RAG collection '$COLLECTION_NAME'..."
    curl -s -X DELETE "$BASE_URL/collections/$COLLECTION_NAME" > /dev/null
    print_success "Collection deleted"
}

# Main execution
main() {
    echo -e "${BLUE}🤖 VittoriaDB RAG System with cURL${NC}"
    echo "===================================="
    echo -e "${YELLOW}Building a Retrieval-Augmented Generation system${NC}"
    
    # Check if bc is available for calculations
    if ! command -v bc &> /dev/null; then
        print_error "bc (calculator) is required for this script"
        print_info "Install with: brew install bc (macOS) or apt-get install bc (Ubuntu)"
        exit 1
    fi
    
    check_connection
    setup_rag_collection
    create_knowledge_base
    
    # Get collection stats
    stats=$(curl -s "$BASE_URL/stats")
    vectors=$(echo "$stats" | jq '.total_vectors' 2>/dev/null || echo "unknown")
    print_success "RAG system ready with $vectors vectors in knowledge base"
    
    demonstrate_rag_queries
    analyze_rag_performance
    
    # Interactive demo
    echo ""
    read -p "$(echo -e ${YELLOW}Would you like to try the interactive RAG demo? [y/N]: ${NC})" -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        interactive_rag_demo
    fi
    
    cleanup
    
    echo -e "\n${GREEN}🎉 RAG example completed successfully!${NC}"
    echo -e "\n${YELLOW}Key Features Demonstrated:${NC}"
    echo "- Knowledge base creation and chunking"
    echo "- Text embedding generation (simplified)"
    echo "- Semantic search and retrieval"
    echo "- Context-aware answer generation"
    echo "- Filtered queries by category"
    echo "- Performance analysis and optimization"
    
    echo -e "\n${YELLOW}Production Improvements:${NC}"
    echo "- Use proper embedding models (Sentence Transformers, OpenAI)"
    echo "- Integrate with real language models (GPT, Claude, local LLMs)"
    echo "- Implement advanced chunking strategies"
    echo "- Add query expansion and reranking"
    echo "- Implement caching for frequently asked questions"
}

# Check if jq is available (recommended for JSON processing)
if ! command -v jq &> /dev/null; then
    print_info "jq is not installed. JSON processing will be limited (install with: brew install jq)"
fi

# Run main function
main "$@"
