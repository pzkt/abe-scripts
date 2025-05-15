use rabe::schemes::ac17::*;
use rabe::utils::policy::pest::PolicyLanguage;

#[test]
fn main(){
    let (pk, msk) = setup();
    let plaintext = String::from("our plaintext!").into_bytes();
    let policy = String::from(r#""attribute_0""#);
    let ct: Ac17CpCiphertext =  cp_encrypt(&pk, &policy, &plaintext, PolicyLanguage::HumanPolicy).unwrap();
    let sk: Ac17CpSecretKey = cp_keygen(&msk, &vec!["attribute_0", "attribute_1", "attribute_2", "attribute_3", "attribute_4", "attribute_5", "attribute_6", "attribute_7", "attribute_8", "attribute_9", "attribute_10", "attribute_11", "attribute_12", "attribute_13", "attribute_14", "attribute_15", "attribute_16", "attribute_17", "attribute_18", "attribute_19"]).unwrap();
    assert_eq!(cp_decrypt(&sk, &ct).unwrap(), plaintext);
}