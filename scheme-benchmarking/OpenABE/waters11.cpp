#include <iostream>
#include <string>
#include <cassert>
#include <numeric>
#include <openabe/openabe.h>
#include <openabe/zsymcrypto.h>
#include <benchmark/benchmark.h>

using namespace std;
using namespace oabe;
using namespace oabe::crypto;

std::string build_attributes(size_t count, const std::string& separator) {
    if (count == 0) return "";
    if (count == 1) return "attribute0";
    
    std::vector<std::string> attributes;
    attributes.reserve(count);
    
    for (size_t i = 0; i < count; ++i) {
        attributes.push_back("attribute" + std::to_string(i));
    }
    
    return std::accumulate(
        std::next(attributes.begin()), 
        attributes.end(), 
        attributes[0],
        [&separator](const std::string& a, const std::string& b) {
            return a + separator + b;
        }
    );
}

std::string get_msg() {
    return "wow schgloopy!";
}

static void setup(benchmark::State& state){
    for (auto _: state){
        InitializeOpenABE();
	OpenABECryptoContext cpabe("CP-ABE");
	cpabe.generateParams();
        ShutdownOpenABE();
    }
}

static void keygen(benchmark::State& state){
    
    const size_t attribute_count = state.range(0);
    InitializeOpenABE();
    OpenABECryptoContext cpabe("CP-ABE");
    cpabe.generateParams();

    const std::string attributes = build_attributes(attribute_count,"|");

    for (auto _ : state) {
        cpabe.keygen(attributes, "key0");
    }
    ShutdownOpenABE();

    state.SetComplexityN(attribute_count);
    state.counters["Attributes"] = attribute_count;
}

static void and_encrypt(benchmark::State& state){
    const size_t attribute_count = state.range(0);
    InitializeOpenABE();
    OpenABECryptoContext cpabe("CP-ABE");
    cpabe.generateParams();

    const std::string attributes = build_attributes(attribute_count," and ");

    string ct, pt1 = get_msg(), pt2;

    for (auto _ : state) {
        cpabe.encrypt(attributes, pt1, ct);
    }
    ShutdownOpenABE();

    state.SetComplexityN(attribute_count);
    state.counters["Attributes"] = attribute_count;
}

static void or_encrypt(benchmark::State& state){
    const size_t attribute_count = state.range(0);
    InitializeOpenABE();
    OpenABECryptoContext cpabe("CP-ABE");
    cpabe.generateParams();

    const std::string attributes = build_attributes(attribute_count," or ");

    string ct, pt1 = get_msg(), pt2;

    for (auto _ : state) {
        cpabe.encrypt(attributes, pt1, ct);
    }
    ShutdownOpenABE();

    state.SetComplexityN(attribute_count);
    state.counters["Attributes"] = attribute_count;
}

static void and_decrypt(benchmark::State& state){
    const size_t attribute_count = state.range(0);
    InitializeOpenABE();
    OpenABECryptoContext cpabe("CP-ABE");
    cpabe.generateParams();

    const std::string policy_attr = build_attributes(attribute_count," and ");
    string ct, pt1 = get_msg(), pt2;
    cpabe.encrypt(policy_attr, pt1, ct);

    const std::string key_attr = build_attributes(attribute_count,"|");
    cpabe.keygen(key_attr, "key0");

    for (auto _ : state) {
        bool result = cpabe.decrypt("key0", ct, pt2);
  	assert(result && pt1 == pt2);
    }
    ShutdownOpenABE();

    state.SetComplexityN(attribute_count);
    state.counters["Attributes"] = attribute_count;
}

static void or_decrypt(benchmark::State& state){
    const size_t attribute_count = state.range(0);
    InitializeOpenABE();
    OpenABECryptoContext cpabe("CP-ABE");
    cpabe.generateParams();

    const std::string policy_attr = build_attributes(attribute_count," or ");
    string ct, pt1 = get_msg(), pt2;
    cpabe.encrypt(policy_attr, pt1, ct);

    cpabe.keygen("attribute" + std::to_string(attribute_count-1), "key0");

    for (auto _ : state) {
        bool result = cpabe.decrypt("key0", ct, pt2);
  	assert(result && pt1 == pt2);
    }
    ShutdownOpenABE();

    state.SetComplexityN(attribute_count);
    state.counters["Attributes"] = attribute_count;
}


static void AttributeRange(benchmark::internal::Benchmark* b) {
    b->Args({1});
    for (int i=5; i <=50; i+=5){
	b->Args({i});
    }
}

BENCHMARK(setup)->Complexity(benchmark::oN);
BENCHMARK(keygen)->Apply(AttributeRange)->Complexity(benchmark::oN);
BENCHMARK(and_encrypt)->Apply(AttributeRange)->Complexity(benchmark::oN);
BENCHMARK(or_encrypt)->Apply(AttributeRange)->Complexity(benchmark::oN);
BENCHMARK(and_decrypt)->Apply(AttributeRange)->Complexity(benchmark::oN);
BENCHMARK(or_decrypt)->Apply(AttributeRange)->Complexity(benchmark::oN);

BENCHMARK_MAIN();
