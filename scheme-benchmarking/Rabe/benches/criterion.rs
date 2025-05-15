use criterion::criterion_main;

mod bdabe_bench;
mod fame_bench;
mod ghw11_bench;
mod mke08_bench;
mod bsw07_bench;

criterion_main! {
    fame_bench::benches,
    bsw07_bench::benches
}