use criterion::{black_box, criterion_group, BenchmarkId, Criterion, SamplingMode, Throughput};

use rabe::schemes::bsw::*;
use rabe::utils::policy::pest::PolicyLanguage;

pub const REPEATS:usize = 10;
pub const ATTR_NR: [i32; 11] = [1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50];
pub const PLAINTEXT: &str = "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul";

pub fn setup_bench(c: &mut Criterion){
    c.bench_function("bsw07_setup", |b| b.iter(|| setup()));
}

pub fn keygen_bench(c: &mut Criterion){
    let (pk, msk) = setup();

    let mut group = c.benchmark_group("bsw07_keygen_attributes");
    for &a in ATTR_NR.iter() {
        let attributes: Vec<String> = (0..a).map(|i| format!("attribute{}", i)).collect();
        let attr_refs: Vec<&str> = attributes.iter().map(|s| s.as_str()).collect();

        //sanity check
        assert!(keygen(&pk, &msk, &attr_refs).is_some());

        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &attr_refs,
            |b, attr_refs| { b.iter(|| keygen(&pk, &msk, &attr_refs))});
    }
}

pub fn delegate_bench(c: &mut Criterion){
    let (pk, msk) = setup();

    let mut group = c.benchmark_group("bsw07_delegate_attributes");
    for &a in ATTR_NR.iter() {
        let attributes: Vec<String> = (0..a).map(|i| format!("attribute{}", i)).collect();
        let attr_refs: Vec<&str> = attributes.iter().map(|s| s.as_str()).collect();
        let delegated_attr_refs: &[&str] = &attr_refs[0..attr_refs.len()];

        let sk = keygen(&pk, &msk, &attr_refs).unwrap();

        assert!(delegate(&pk, &sk, delegated_attr_refs).is_some());

        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &delegated_attr_refs,
            |b, delegated_attr_refs| { b.iter(|| delegate(&pk, &sk, delegated_attr_refs))});
    }
}

pub fn and_encryption_bench(c: &mut Criterion){
    let (pk, _) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("bsw07_and_encryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        //sanity check
        assert!(encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext).is_ok());

        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &policy,
            |b, policy| { b.iter(|| {encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext)})});
    }
}

pub fn or_encryption_bench(c: &mut Criterion){
    let (pk, _) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("bsw07_or_encryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute_0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        //sanity check
        assert!(encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext).is_ok());

        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &policy,
            |b, policy| { b.iter(|| {encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext)})});
    }
}

pub fn and_decryption_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("bsw07_and_decryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" and ("attribute{}""#, &i)); }
            policy.push_str(&format!(r#" and "attribute{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        let ct = encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext).unwrap();

        let attributes: Vec<String> = (0..a).map(|i| format!("attribute{}", i)).collect();
        let attr_refs: Vec<&str> = attributes.iter().map(|s| s.as_str()).collect();

        let sk = keygen(&pk, &msk, &attr_refs).unwrap();

        //sanity check
        assert!(decrypt(&sk, &ct).is_ok());

        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &ct,
            |b, ct| { b.iter(|| {decrypt(&sk, &ct)})});
    }
}

pub fn or_decryption_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("bsw07_or_decryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        let ct = encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext).unwrap();
        let sk = keygen(&pk, &msk, &vec![format!("attribute{}", a-1).as_str()]).unwrap();

        //sanity check
        assert!(decrypt(&sk, &ct).is_ok());

        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &ct,
            |b, ct| { b.iter(|| {decrypt(&sk, &ct)})});
    }
}

criterion_group! {
    name = benches;
    config = Criterion::default().sample_size(REPEATS);
    targets = setup_bench, keygen_bench, delegate_bench, and_encryption_bench, or_encryption_bench, and_decryption_bench, or_decryption_bench
}