#[cfg(test)]
mod test {
    use rabe::schemes::bsw::*;
    use rabe::utils::policy::pest::PolicyLanguage;

    #[test]
    fn main(){
        let (pk, msk) = setup();
        let plaintext = String::from("dance like no one's watching, encrypt like everyone is!").into_bytes();
        let policy = String::from(r#""attribute0""#);
        let ct_cp: CpAbeCiphertext = encrypt(&pk, &policy, PolicyLanguage::HumanPolicy, &plaintext).unwrap();
        let sk: CpAbeSecretKey = keygen(&pk, &msk, &vec!["attribute0"]).unwrap();
        assert_eq!(decrypt(&sk, &ct_cp).unwrap(), plaintext);
    }
}