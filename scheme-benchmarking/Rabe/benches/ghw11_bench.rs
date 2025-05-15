use criterion::{criterion_group, BenchmarkId, Criterion, SamplingMode, Throughput};

use rabe::schemes::ghw11::*;
use rabe::utils::policy::pest::PolicyLanguage;

pub const REPEATS:usize = 10;
pub const ATTR_NR: [i32; 11] = [1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50];
pub const PLAINTEXT: &str = "Blue canary in the outlet by the light switch, who watches over you. Make a little birdhouse in your soul";

pub fn setup_bench(c: &mut Criterion){
    c.bench_function("fame_setup", |b| b.iter(|| setup()));
}

pub fn keygen_bench(c: &mut Criterion){
    let (pk, msk) = setup();

    let mut group = c.benchmark_group("ghw11_keygen_attributes");
    for &a in ATTR_NR.iter() {
        let attributes: Vec<String> = (0..a).map(|i| format!("attribute{}", i)).collect();

        //sanity check
        assert!(keygen(&pk, &msk, &attributes).is_some());
        
        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &attributes,
            |b, attributes| { b.iter(|| keygen(&pk, &msk, &attributes))});
    }
}

pub fn transform_keygen_bench(c: &mut Criterion){
    let (pk, msk) = setup();

    let mut group = c.benchmark_group("ghw11_transform_keygen_attributes");
    for &a in ATTR_NR.iter() {
        let attributes: Vec<String> = (0..a).map(|i| format!("attribute{}", i)).collect();

        let sk = keygen(&pk, &msk, &attributes).unwrap();

        //sanity check
        assert!(tkgen(sk.clone()).is_some());
        
        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &sk,
            |b, sk| { b.iter(|| {tkgen(sk.clone()).is_some()})});
    }
}

pub fn and_encryption_bench(c: &mut Criterion){
    let (pk, _) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("ghw11_and_encryption_attributes");
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

    let mut group = c.benchmark_group("ghw11_or_encryption_attributes");
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

pub fn and_transform_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("ghw11_and_transform_attributes");
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

        let sk = keygen(&pk, &msk, &attributes).unwrap();
        let (tk, _) = tkgen(sk).unwrap();

        //sanity check
        assert!(transform(ct.clone(), tk.clone()).is_ok());
        
        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &ct,
            |b, ct| { b.iter(|| {transform(ct.clone(), tk.clone())})});
    }
}

pub fn or_transform_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("ghw11_or_transform_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        let ct = encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext).unwrap();

        let attributes: Vec<String> = (0..a).map(|i| format!("attribute{}", i)).collect();

        let sk = keygen(&pk, &msk, &attributes).unwrap();
        let (tk, _) = tkgen(sk).unwrap();

        //sanity check
        assert!(transform(ct.clone(), tk.clone()).is_ok());
        
        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &ct,
            |b, ct| { b.iter(|| {transform(ct.clone(), tk.clone())})});
    }
}

pub fn and_decryption_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("ghw11_and_decryption_attributes");
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

        let sk = keygen(&pk, &msk, &attributes).unwrap();
        let (tk, rk) = tkgen(sk).unwrap();

        let tct = transform(ct.clone(), tk).unwrap();

        //sanity check
        assert!(decrypt_out(tct.clone(), rk.clone(), ct.data.clone()).is_ok());
        
        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &tct,
            |b, tct| { b.iter(|| {decrypt_out(tct.clone(), rk.clone(), ct.data.clone())})});
    }
}

pub fn or_decryption_bench(c: &mut Criterion){
    let (pk, msk) = setup();
    let plaintext = String::from(PLAINTEXT).into_bytes();

    let mut group = c.benchmark_group("ghw11_or_decryption_attributes");
    for &a in ATTR_NR.iter() {
        let mut policy = r#""attribute0""#.to_owned();
        if a != 1 {
            policy = r#"("attribute0""#.to_owned();
            for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute{}""#, &i)); }
            policy.push_str(&format!(r#" or "attribute{}""#, a-1));
            for _ in 1..a { policy.push_str(")"); }
        }

        let ct = encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext).unwrap();

        let attributes: Vec<String> = (0..a).map(|i| format!("attribute{}", i)).collect();

        let sk = keygen(&pk, &msk, &attributes).unwrap();
        let (tk, rk) = tkgen(sk).unwrap();

        let tct = transform(ct.clone(), tk).unwrap();

        //sanity check
        assert!(decrypt_out(tct.clone(), rk.clone(), ct.data.clone()).is_ok());
        
        group.sampling_mode(SamplingMode::Flat);
        group.throughput(Throughput::Elements(a as u64));
        group.bench_with_input(
            BenchmarkId::from_parameter(a),
            &tct,
            |b, tct| { b.iter(|| {decrypt_out(tct.clone(), rk.clone(), ct.data.clone())})});
    }
}

criterion_group! {
    name = benches;
    config = Criterion::default().sample_size(REPEATS);
    targets = setup_bench, keygen_bench, transform_keygen_bench, and_encryption_bench, or_encryption_bench, and_transform_bench, or_transform_bench, and_decryption_bench, or_decryption_bench
}