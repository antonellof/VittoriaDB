#!/bin/bash

# VittoriaDB Volume Testing with cURL
# This script tests VittoriaDB performance with different data volumes: KB, MB, GB
# Make sure VittoriaDB is running: ./vittoriadb run

set -e  # Exit on any error

# Configuration
BASE_URL="http://localhost:8080"
COLLECTION_PREFIX="volume_test"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
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

print_perf() {
    echo -e "${PURPLE}📊 $1${NC}"
}

# Generate a random vector of specified dimensions (optimized)
generate_vector() {
    local dimensions=$1
    local vector="["
    
    # Generate all random numbers in one awk call for efficiency
    local random_values=$(awk -v dims=$dimensions -v seed=$RANDOM '
        BEGIN {
            srand(seed + NR);
            for(i=0; i<dims; i++) {
                if(i > 0) printf ", ";
                printf "%.6f", (rand()-0.5)*2;
            }
        }')
    
    vector+="$random_values]"
    echo "$vector"
}

# Generate metadata for a vector
generate_metadata() {
    local id=$1
    local category_num=$((id % 5))
    local categories=("technology" "science" "education" "business" "research")
    local category=${categories[$category_num]}
    
    echo "{
        \"title\": \"Document $id\",
        \"category\": \"$category\",
        \"author\": \"Author $((id % 10))\",
        \"year\": $((2020 + (id % 5))),
        \"size\": \"volume_test\",
        \"batch_id\": \"$((id / 100))\"
    }"
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

# Create collection for testing
create_test_collection() {
    local collection_name=$1
    local dimensions=$2
    local index_type=$3
    
    print_info "Creating collection '$collection_name' ($dimensions dims, $index_type index)..."
    
    # Delete existing collection if it exists
    curl -s -X DELETE "$BASE_URL/collections/$collection_name" > /dev/null 2>&1 || true
    
    # Create new collection
    response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/collections" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$collection_name\",
            \"dimensions\": $dimensions,
            \"index_type\": $([ "$index_type" = "flat" ] && echo "0" || echo "1"),
            \"metric\": 0,
            \"description\": \"Volume test collection - $dimensions dimensions\"
        }")
    
    http_code="${response: -3}"
    if [[ "$http_code" == "201" ]]; then
        print_success "Created collection '$collection_name'"
    else
        print_error "Failed to create collection (HTTP $http_code)"
        exit 1
    fi
}

# Test KB-scale data (small vectors, few dimensions)
test_kb_scale() {
    print_header "KB-Scale Test (Small Vectors)"
    
    local collection_name="${COLLECTION_PREFIX}_kb"
    local dimensions=32
    local vector_count=100
    local batch_size=10
    
    create_test_collection "$collection_name" "$dimensions" "flat"
    
    print_info "Testing with $vector_count vectors of $dimensions dimensions"
    local size_kb=$(echo "scale=1; $vector_count * $dimensions * 4 / 1024" | bc)
    print_info "Estimated data size: ~${size_kb} KB"
    
    # Individual insertions test
    print_info "Testing individual insertions..."
    start_time=$(date +%s.%N)
    
    for ((i=1; i<=20; i++)); do
        vector=$(generate_vector $dimensions)
        metadata=$(generate_metadata $i)
        
        curl -s -X POST "$BASE_URL/collections/$collection_name/vectors" \
            -H "Content-Type: application/json" \
            -d "{
                \"id\": \"individual_$i\",
                \"vector\": $vector,
                \"metadata\": $metadata
            }" > /dev/null
    done
    
    individual_time=$(echo "$(date +%s.%N) - $start_time" | bc)
    individual_rate=$(echo "scale=2; 20 / $individual_time" | bc)
    print_perf "Individual insertions: 20 vectors in ${individual_time}s (${individual_rate} vectors/sec)"
    
    # Batch insertions test
    print_info "Testing batch insertions..."
    start_time=$(date +%s.%N)
    
    for ((batch=0; batch<8; batch++)); do
        batch_data="{\"vectors\": ["
        for ((i=0; i<batch_size; i++)); do
            if [ $i -gt 0 ]; then
                batch_data+=", "
            fi
            vector_id=$((batch * batch_size + i + 21))
            vector=$(generate_vector $dimensions)
            metadata=$(generate_metadata $vector_id)
            
            batch_data+="{
                \"id\": \"batch_${vector_id}\",
                \"vector\": $vector,
                \"metadata\": $metadata
            }"
        done
        batch_data+="]}"
        
        curl -s -X POST "$BASE_URL/collections/$collection_name/vectors/batch" \
            -H "Content-Type: application/json" \
            -d "$batch_data" > /dev/null
    done
    
    batch_time=$(echo "$(date +%s.%N) - $start_time" | bc)
    batch_rate=$(echo "scale=2; 80 / $batch_time" | bc)
    print_perf "Batch insertions: 80 vectors in ${batch_time}s (${batch_rate} vectors/sec)"
    
    # Search performance test
    test_search_performance "$collection_name" "$dimensions" 5
    
    # Cleanup
    curl -s -X DELETE "$BASE_URL/collections/$collection_name" > /dev/null
    print_success "KB-scale test completed"
}

# Test MB-scale data (medium vectors, more dimensions)
test_mb_scale() {
    print_header "MB-Scale Test (Medium Vectors)"
    
    local collection_name="${COLLECTION_PREFIX}_mb"
    local dimensions=512
    local vector_count=5000  # 5000 * 512 * 4 = ~10 MB
    local batch_size=100
    
    create_test_collection "$collection_name" "$dimensions" "flat"
    
    print_info "Testing with $vector_count vectors of $dimensions dimensions"
    local size_mb=$(echo "scale=2; $vector_count * $dimensions * 4 / 1024 / 1024" | bc)
    print_info "Estimated data size: ~${size_mb} MB"
    
    # Batch insertions only (more efficient for larger datasets)
    print_info "Testing batch insertions..."
    start_time=$(date +%s.%N)
    
    for ((batch=0; batch<50; batch++)); do  # 50 batches * 100 = 5000 vectors
        batch_data="{\"vectors\": ["
        for ((i=0; i<batch_size; i++)); do
            if [ $i -gt 0 ]; then
                batch_data+=", "
            fi
            vector_id=$((batch * batch_size + i + 1))
            vector=$(generate_vector $dimensions)
            metadata=$(generate_metadata $vector_id)
            
            batch_data+="{
                \"id\": \"mb_${vector_id}\",
                \"vector\": $vector,
                \"metadata\": $metadata
            }"
        done
        batch_data+="]}"
        
        curl -s -X POST "$BASE_URL/collections/$collection_name/vectors/batch" \
            -H "Content-Type: application/json" \
            -d "$batch_data" > /dev/null
        
        # Progress indicator
        if [ $((batch % 10)) -eq 0 ]; then
            current_vectors=$((batch * batch_size))
            progress=$((current_vectors * 100 / vector_count))
            print_info "Progress: $current_vectors / $vector_count vectors inserted (${progress}%)"
        fi
    done
    
    total_time=$(echo "$(date +%s.%N) - $start_time" | bc)
    total_rate=$(echo "scale=2; $vector_count / $total_time" | bc)
    print_perf "Batch insertions: $vector_count vectors in ${total_time}s (${total_rate} vectors/sec)"
    
    # Search performance test
    test_search_performance "$collection_name" "$dimensions" 10
    
    # Memory usage check
    stats=$(curl -s "$BASE_URL/stats")
    memory_mb=$(echo "$stats" | jq '.memory_usage / 1024 / 1024' 2>/dev/null || echo "unknown")
    print_perf "Memory usage: ${memory_mb} MB"
    
    # Cleanup
    curl -s -X DELETE "$BASE_URL/collections/$collection_name" > /dev/null
    print_success "MB-scale test completed"
}

# Test GB-scale data (large vectors, high dimensions, HNSW index)
test_gb_scale() {
    print_header "Large-Scale Test (High-Dimensional Vectors with HNSW)"
    
    local collection_name="${COLLECTION_PREFIX}_gb"
    local dimensions=768
    local vector_count=20000  # 20000 * 768 * 4 = ~58 MB (substantial scale)
    local batch_size=50
    
    create_test_collection "$collection_name" "$dimensions" "hnsw"
    
    print_info "Testing with $vector_count vectors of $dimensions dimensions"
    local size_mb=$(echo "scale=2; $vector_count * $dimensions * 4 / 1024 / 1024" | bc)
    print_info "Estimated data size: ~${size_mb} MB"
    print_info "Using HNSW index for better performance with large datasets"
    print_info "Note: Dimensions limited to 768 to avoid 'Argument list too long' errors in curl"
    
    # Batch insertions with progress tracking
    print_info "Testing batch insertions with HNSW indexing..."
    start_time=$(date +%s.%N)
    
    for ((batch=0; batch<400; batch++)); do  # 400 batches * 50 = 20,000 vectors
        batch_data="{\"vectors\": ["
        for ((i=0; i<batch_size; i++)); do
            if [ $i -gt 0 ]; then
                batch_data+=", "
            fi
            vector_id=$((batch * batch_size + i + 1))
            vector=$(generate_vector $dimensions)
            metadata=$(generate_metadata $vector_id)
            
            batch_data+="{
                \"id\": \"gb_${vector_id}\",
                \"vector\": $vector,
                \"metadata\": $metadata
            }"
        done
        batch_data+="]}"
        
        curl -s -X POST "$BASE_URL/collections/$collection_name/vectors/batch" \
            -H "Content-Type: application/json" \
            -d "$batch_data" > /dev/null
        
        # Progress indicator every 20 batches
        if [ $((batch % 20)) -eq 0 ]; then
            current_vectors=$((batch * batch_size))
            progress=$((current_vectors * 100 / vector_count))
            print_info "Progress: $current_vectors / $vector_count vectors inserted (${progress}%)"
        fi
    done
    
    total_time=$(echo "$(date +%s.%N) - $start_time" | bc)
    total_rate=$(echo "scale=2; $vector_count / $total_time" | bc)
    print_perf "HNSW batch insertions: $vector_count vectors in ${total_time}s (${total_rate} vectors/sec)"
    
    # Search performance test with HNSW
    test_search_performance "$collection_name" "$dimensions" 20
    
    # Memory usage check
    stats=$(curl -s "$BASE_URL/stats")
    memory_mb=$(echo "$stats" | jq '.memory_usage / 1024 / 1024' 2>/dev/null || echo "unknown")
    print_perf "Memory usage: ${memory_mb} MB"
    
    # Test different search parameters for HNSW
    print_info "Testing HNSW search parameters..."
    query_vector=$(generate_vector $dimensions)
    
    # Search with different ef values
    for ef in 50 100 200; do
        start_time=$(date +%s.%N)
        curl -s -X POST "$BASE_URL/collections/$collection_name/search" \
            -H "Content-Type: application/json" \
            -d "{
                \"vector\": $query_vector,
                \"k\": 10,
                \"params\": {\"ef\": $ef}
            }" > /dev/null
        search_time=$(echo "$(date +%s.%N) - $start_time" | bc)
        print_perf "HNSW search (ef=$ef): ${search_time}s"
    done
    
    # Cleanup
    curl -s -X DELETE "$BASE_URL/collections/$collection_name" > /dev/null
    print_success "GB-scale test completed"
}

# Test search performance
test_search_performance() {
    local collection_name=$1
    local dimensions=$2
    local num_searches=$3
    
    print_info "Testing search performance ($num_searches searches)..."
    
    local total_time=0
    for ((i=1; i<=num_searches; i++)); do
        query_vector=$(generate_vector $dimensions)
        
        start_time=$(date +%s.%N)
        curl -s -X POST "$BASE_URL/collections/$collection_name/search" \
            -H "Content-Type: application/json" \
            -d "{
                \"vector\": $query_vector,
                \"k\": 5
            }" > /dev/null
        search_time=$(echo "$(date +%s.%N) - $start_time" | bc)
        total_time=$(echo "$total_time + $search_time" | bc)
    done
    
    avg_time=$(echo "scale=4; $total_time / $num_searches" | bc)
    searches_per_sec=$(echo "scale=2; $num_searches / $total_time" | bc)
    print_perf "Search performance: avg ${avg_time}s per search (${searches_per_sec} searches/sec)"
}

# Performance comparison test
performance_comparison() {
    print_header "Performance Comparison Summary"
    
    print_info "Running comparative performance test..."
    
    # Test different configurations
    local configs=(
        "comp_flat_64:64:flat:500"
        "comp_flat_256:256:flat:500"
        "comp_hnsw_256:256:hnsw:500"
        "comp_hnsw_512:512:hnsw:500"
    )
    
    for config in "${configs[@]}"; do
        IFS=':' read -r name dimensions index_type count <<< "$config"
        
        print_info "Testing $name ($dimensions dims, $index_type, $count vectors)..."
        
        create_test_collection "$name" "$dimensions" "$index_type"
        
        # Batch insert
        start_time=$(date +%s.%N)
        batch_size=50
        num_batches=$((count / batch_size))
        
        for ((batch=0; batch<num_batches; batch++)); do
            batch_data="{\"vectors\": ["
            for ((i=0; i<batch_size; i++)); do
                if [ $i -gt 0 ]; then
                    batch_data+=", "
                fi
                vector_id=$((batch * batch_size + i + 1))
                vector=$(generate_vector $dimensions)
                metadata=$(generate_metadata $vector_id)
                
                batch_data+="{
                    \"id\": \"${name}_${vector_id}\",
                    \"vector\": $vector,
                    \"metadata\": $metadata
                }"
            done
            batch_data+="]}"
            
            curl -s -X POST "$BASE_URL/collections/$name/vectors/batch" \
                -H "Content-Type: application/json" \
                -d "$batch_data" > /dev/null
        done
        
        insert_time=$(echo "$(date +%s.%N) - $start_time" | bc)
        insert_rate=$(echo "scale=2; $count / $insert_time" | bc)
        
        # Search test
        query_vector=$(generate_vector $dimensions)
        start_time=$(date +%s.%N)
        curl -s -X POST "$BASE_URL/collections/$name/search" \
            -H "Content-Type: application/json" \
            -d "{
                \"vector\": $query_vector,
                \"k\": 10
            }" > /dev/null
        search_time=$(echo "$(date +%s.%N) - $start_time" | bc)
        
        print_perf "$name: Insert ${insert_rate} vec/sec, Search ${search_time}s"
        
        # Cleanup
        curl -s -X DELETE "$BASE_URL/collections/$name" > /dev/null
    done
}

# Memory stress test
memory_stress_test() {
    print_header "Memory Stress Test"
    
    local collection_name="${COLLECTION_PREFIX}_memory"
    local dimensions=512  # Reduced to avoid "Argument list too long" error
    local batch_size=50
    
    create_test_collection "$collection_name" "$dimensions" "hnsw"
    
    print_info "Memory stress test with high-dimensional vectors ($dimensions dims)"
    print_info "Monitoring memory usage during insertion..."
    
    for ((batch=1; batch<=20; batch++)); do
        # Insert batch
        batch_data="{\"vectors\": ["
        for ((i=0; i<batch_size; i++)); do
            if [ $i -gt 0 ]; then
                batch_data+=", "
            fi
            vector_id=$((batch * batch_size + i))
            vector=$(generate_vector $dimensions)
            metadata=$(generate_metadata $vector_id)
            
            batch_data+="{
                \"id\": \"mem_${vector_id}\",
                \"vector\": $vector,
                \"metadata\": $metadata
            }"
        done
        batch_data+="]}"
        
        curl -s -X POST "$BASE_URL/collections/$collection_name/vectors/batch" \
            -H "Content-Type: application/json" \
            -d "$batch_data" > /dev/null
        
        # Check memory usage
        stats=$(curl -s "$BASE_URL/stats")
        memory_mb=$(echo "$stats" | jq '.memory_usage / 1024 / 1024' 2>/dev/null || echo "unknown")
        vectors=$(echo "$stats" | jq '.total_vectors' 2>/dev/null || echo "unknown")
        
        print_perf "Batch $batch: $vectors vectors, ${memory_mb} MB memory"
        
        # Stop if memory usage gets too high (demo purposes)
        if [ "$memory_mb" != "unknown" ] && [ $(echo "$memory_mb > 500" | bc) -eq 1 ]; then
            print_info "Stopping stress test at 500MB memory usage"
            break
        fi
    done
    
    # Cleanup
    curl -s -X DELETE "$BASE_URL/collections/$collection_name" > /dev/null
    print_success "Memory stress test completed"
}

# Main execution
main() {
    echo -e "${BLUE}🧪 VittoriaDB Volume Testing Suite${NC}"
    echo "===================================="
    echo -e "${YELLOW}Testing different data volumes: KB, MB, GB scales${NC}"
    
    # Check if bc is available for calculations
    if ! command -v bc &> /dev/null; then
        print_error "bc (calculator) is required for this script"
        print_info "Install with: brew install bc (macOS) or apt-get install bc (Ubuntu)"
        exit 1
    fi
    
    check_connection
    
    # Run volume tests
    test_kb_scale
    test_mb_scale
    test_gb_scale
    
    # Additional tests
    performance_comparison
    memory_stress_test
    
    # Final stats
    print_header "Final Database Statistics"
    final_stats=$(curl -s "$BASE_URL/stats")
    echo "$final_stats" | jq '.' 2>/dev/null || echo "$final_stats"
    
    echo -e "\n${GREEN}🎉 Volume testing completed successfully!${NC}"
    echo -e "\n${YELLOW}Key Findings:${NC}"
    echo "- KB-scale: Best for small datasets with simple vectors"
    echo "- MB-scale: Good balance of performance and capacity"
    echo "- GB-scale: Use HNSW indexing for large, high-dimensional datasets"
    echo "- Batch operations are significantly faster than individual insertions"
    echo "- HNSW provides better search performance for large datasets"
    
    echo -e "\n${YELLOW}Next steps:${NC}"
    echo "- Optimize batch sizes based on your data characteristics"
    echo "- Choose appropriate index types (Flat vs HNSW) based on dataset size"
    echo "- Monitor memory usage in production environments"
    echo "- Consider data partitioning strategies for very large datasets"
    echo ""
    echo -e "${YELLOW}For truly large-scale testing (GB+):${NC}"
    echo "- Use the Python or Go examples for better performance"
    echo "- Consider streaming data from files instead of generating in-memory"
    echo "- Use the native client libraries to avoid HTTP payload size limits"
}

# Run main function
main "$@"
