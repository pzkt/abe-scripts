#include <iostream>
#include <fstream>
#include <random>
#include <sstream>
#include <vector>
#include <string>
#include <cassert>
#include <numeric>
#include <openabe/openabe.h>
#include <openabe/zsymcrypto.h>
#include <benchmark/benchmark.h>

using namespace std;
using namespace oabe;
using namespace oabe::crypto;

void update_csv(const std::string& file_name, const std::string& index, const std::string& column, const std::string& value) {
    std::vector<std::vector<std::string>> records;
    std::vector<std::string> headers;
    bool file_exists = true;

    std::ifstream file(file_name);
    if (file.is_open()) {
        std::string line;
        while (std::getline(file, line)) {
            std::stringstream ss(line);
            std::string field;
            std::vector<std::string> row;
            while (std::getline(ss, field, ',')) {
                row.push_back(field);
            }
            records.push_back(row);
        }
        if (!records.empty()) {
            headers = records[0];
        }
        file.close();
    } else {
        file_exists = false;
    }

    if (records.empty()) {
        headers = {"attributes"};
        records.push_back(headers);
    }

    int column_index = -1;
    for (size_t i = 0; i < headers.size(); ++i) {
        if (headers[i] == column) {
            column_index = i;
            break;
        }
    }
    if (column_index == -1) {
        headers.push_back(column);
        column_index = headers.size() - 1;
        records[0] = headers;
        
        for (size_t i = 1; i < records.size(); ++i) {
            records[i].resize(column_index + 1, "");
        }
    }

    int row_index = -1;
    for (size_t i = 1; i < records.size(); ++i) {
        if (records[i][0] == index) {
            row_index = i;
            break;
        }
    }

    if (row_index == -1) {
        std::vector<std::string> new_row(headers.size(), "");
        new_row[0] = index;
        records.push_back(new_row);
        row_index = records.size() - 1;
    }

    records[row_index].resize(column_index + 1, "");
    records[row_index][column_index] = value;

    std::ofstream outfile(file_name);
    if (outfile.is_open()) {
        for (const auto& row : records) {
            for (size_t i = 0; i < row.size(); ++i) {
                outfile << row[i];
                if (i != row.size() - 1) {
                    outfile << ",";
                }
            }
            outfile << "\n";
        }
        outfile.close();
    } else {
        std::cerr << "Error writing to file" << std::endl;
    }
}


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
    return "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul";
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
	cout << ct.size() << endl;
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

//BENCHMARK_MAIN();

// --------------- size benchmark ---------------

//generate a random string with a given byte size
std::string random_msg(size_t size) {
    std::random_device rd;
    std::vector<uint8_t> buffer(size);

    // Generate random bytes and store them in the buffer
    for (size_t i = 0; i < size; ++i) {
        buffer[i] = static_cast<uint8_t>(rd() & 0xFF);
    }

    // Convert the vector of bytes to a string
    return std::string(buffer.begin(), buffer.end());
}

int main(){
    //update_csv("example.csv", "123", "new_column", "new_value");
    
    for(int i=0; i < 25; i++){
        size_t size = 1 << i;
        int attribute_counts[] = {1,5,10,15,20,25,30,35,40,45,50};
        InitializeOpenABE();
        OpenABECryptoContext cpabe("CP-ABE");
        cpabe.generateParams();

        for (int count : attribute_counts){

            const std::string attributes = build_attributes(count," and ");

            string ct, pt1 = random_msg(size);

            cpabe.encrypt(attributes, pt1, ct);
            update_csv("openabe_waters11_ct.csv", std::to_string(count), "hybrid " + std::to_string(1 << i), std::to_string(ct.size()));
        }
    	ShutdownOpenABE();
    }
}
