#[cfg(test)]
mod test {
    use rabe::schemes::bsw::*;
    use rabe::utils::policy::pest::PolicyLanguage;
    use rand::Rng;
    use bincode;
    use crate::shared::*;


    pub const ATTR_NR: [i32; 11] = [1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50];

    #[test]
    fn main(){
        let (pk, _) = setup();

        for n in 0..25{
            let mut content = vec![0u8; 1usize << n];
            rand::thread_rng().fill(&mut content[..]);

            for &a in ATTR_NR.iter() {
                let mut policy = r#""attribute0""#.to_owned();
                if a != 1 {
                    policy = r#"("attribute0""#.to_owned();
                    for i in 1..(a-1) { policy.push_str(&format!(r#" or ("attribute{}""#, &i)); }
                    policy.push_str(&format!(r#" or "attribute{}""#, a-1));
                    for _ in 1..a { policy.push_str(")"); }
                }
            
                let ct = encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &content).unwrap();
                let serialized_ct = bincode::serialize(&ct).unwrap();
                // horrible string and &str conversions
                _ = update_csv("rabe_bsw07_ct.csv", &a.to_string(), &("single ".to_owned() + &(1usize << n).to_string()), &(serialized_ct.len()).to_string());            
            }
        }
    }
}