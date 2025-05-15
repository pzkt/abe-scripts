use criterion::{black_box, criterion_group, BenchmarkId, Criterion, Throughput};

use rabe::schemes::ac17::*;
use rabe::utils::policy::pest::PolicyLanguage;

pub const REPEATS:usize = 10;
pub const ATTR_NR: [i32; 11] = [1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50];
pub const PLAINTEXT: &str = "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul";

pub fn setup_bench(c: &mut Criterion){
    c.bench_function("fame_setup", |b| b.iter(|| setup()));
}

pub fn keygen_bench(c: &mut Criterion){
    let (_, msk) = setup();

    let mut group = c.benchmark_group("fame_keygen_attributes");
    for &a in ATTR_NR.iter() {
        let attributes: Vec<String> = (0..a).map(|i| format!("attribute_{}", i)).collect();
        let attr_refs: Vec<&str> = attributes.iter().map(|s| s.as_str()).collect();
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &attr_refs,
            |b, attr_refs| { b.iter(|| cp_keygen(&msk, black_box(&attr_refs)))});
    }
}

pub fn and_encryption_bench(c: &mut Criterion){
    let (pk, _) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("fame_and_encryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute_0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute_0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute_{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute_{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &policy,
            |b, policy| { b.iter(|| {cp_encrypt(&pk, &policy, &plaintext, PolicyLanguage::HumanPolicy)})});
    }
}

pub fn or_encryption_bench(c: &mut Criterion){
    let (pk, _) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("fame_or_encryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute_0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute_0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute_{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute_{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &policy,
            |b, policy| { b.iter(|| {cp_encrypt(&pk, &policy, &plaintext, PolicyLanguage::HumanPolicy)})});
    }
}

pub fn and_decryption_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("fame_and_decryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute_0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute_0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" and ("attribute_{}""#, &i)); }
            policy.push_str(&format!(r#" and "attribute_{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        let ct = cp_encrypt(&pk, &policy, &plaintext, PolicyLanguage::HumanPolicy).unwrap();

        let attributes: Vec<String> = (0..a).map(|i| format!("attribute_{}", i)).collect();
        let attr_refs: Vec<&str> = attributes.iter().map(|s| s.as_str()).collect();

        let sk = cp_keygen(&msk, &attr_refs).unwrap();

        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &ct,
            |b, ct| { b.iter(|| {cp_decrypt(&sk, &ct)})});
    }
}

pub fn or_decryption_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("fame_or_decryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute_0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute_0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute_{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute_{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        let ct = cp_encrypt(&pk, &policy, &plaintext, PolicyLanguage::HumanPolicy).unwrap();

        let attributes: Vec<String> = (0..a).map(|i| format!("attribute_{}", i)).collect();
        let attr_refs: Vec<&str> = attributes.iter().map(|s| s.as_str()).collect();

        let sk = cp_keygen(&msk, &vec![format!("attribute_{}", a-1).as_str()]).unwrap();

        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &ct,
            |b, ct| { b.iter(|| {cp_decrypt(&sk, &ct)})});
    }
}

criterion_group! {
    name = benches;
    config = Criterion::default().sample_size(REPEATS);
    targets = setup_bench, keygen_bench, and_encryption_bench, or_encryption_bench, and_decryption_bench, or_decryption_bench
}