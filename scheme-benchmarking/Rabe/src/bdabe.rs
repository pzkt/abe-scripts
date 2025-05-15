#[cfg(test)]
mod test {
use rabe::schemes::bdabe::*;
use rabe::utils::policy::pest::PolicyLanguage;

#[test]
fn main(){
    let (pk, msk) = setup();
    let auth1 = authgen(&pk, &msk, &String::from("auth1"));
    let mut sk = keygen(&pk, &auth1, &String::from("u1"));
    let attr_a_pk = request_attribute_pk(&pk, &auth1, "auth1::A").unwrap();
    sk.sk_a.push(request_attribute_sk(&sk.pk, &auth1, "auth1::A").unwrap());
    let plaintext = String::from("our plaintext!").into_bytes();
    let policy = String::from(r#"("auth1::A" or "auth1::B") and "auth1::C""#);
    let ct: BdabeCiphertext = encrypt(&pk, &vec![&attr_a_pk], &policy, PolicyLanguage::HumanPolicy, &plaintext).unwrap();
    let ct_decrypted = decrypt(&sk, &ct);
    assert_eq!(ct_decrypted.is_ok(), true);
    assert_eq!(ct_decrypted.unwrap(), plaintext);
}
}