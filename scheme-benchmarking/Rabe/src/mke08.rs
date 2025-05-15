#[cfg(test)]
mod test {
    use rabe::schemes::mke08::*;
    use rabe::utils::policy::pest::PolicyLanguage;
    
    #[test]
    fn main(){
        let (_pk, _msk) = setup();
        let mut sk = keygen(&_pk, &_msk, "user1");
        let _att1 = "aa1::A";
        let _att2 = "aa2::B";
        let _a1_key = authgen("aa1");
        let _a2_key = authgen("aa2");
        let _att1_pk = request_authority_pk(&_pk, &_att1, &_a1_key).unwrap();
        let _att2_pk = request_authority_pk(&_pk, &_att2, &_a2_key).unwrap();
        sk.sk_a.push(request_authority_sk(&sk.pk, &_att1, &_a1_key).unwrap());
        sk.sk_a.push(request_authority_sk(&sk.pk, &_att2, &_a2_key).unwrap());
        let plaintext = String::from("our plaintext!").into_bytes();
        let policy = String::from(r#"("aa1::A" and "aa2::B") or ("aa2::C") or ("aa2::D")"#);
        let attr_vec: Vec<&Mke08PublicAttributeKey> = vec!(&_att1_pk, &_att2_pk);
        let _ct: Mke08Ciphertext = encrypt(&_pk, &attr_vec.as_slice(), &policy, PolicyLanguage::HumanPolicy, &plaintext).unwrap();
        assert_eq!(decrypt(&sk, &_ct).unwrap(), plaintext);
    }
}