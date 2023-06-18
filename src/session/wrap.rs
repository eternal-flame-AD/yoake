use aes_gcm::{
    aead::{Aead, OsRng},
    AeadCore, Aes256Gcm, Key, KeyInit, Nonce,
};
use base64::{engine::general_purpose, Engine};
use serde::{de::DeserializeOwned, Serialize};

pub fn wrap_json<T: Serialize>(data: T, key: &[u8]) -> String {
    let nonce = Aes256Gcm::generate_nonce(&mut OsRng);
    let key = Key::<Aes256Gcm>::from_slice(key);
    let cipher = Aes256Gcm::new(&key);

    let b64_engine = general_purpose::STANDARD;

    let plaintext = serde_json::to_string(&data).unwrap();
    let ciphertext = cipher
        .encrypt(&nonce, plaintext.as_bytes())
        .expect("Failed to encrypt");

    let mut output = String::new();

    output.push_str(&b64_engine.encode(&nonce));
    output.push_str(":");
    b64_engine.encode_string(&ciphertext, &mut output);

    output
}

pub fn unwrap_json<T: DeserializeOwned>(data: &str, key: &[u8]) -> Option<T> {
    let data = data.splitn(2, ':').collect::<Vec<_>>();
    let nonce_b64 = data.get(0)?;
    let ciphertext_b64 = data.get(1)?;

    let nonce = general_purpose::STANDARD.decode(nonce_b64).ok()?;
    let ciphertext = general_purpose::STANDARD.decode(ciphertext_b64).ok()?;

    let cipher = Aes256Gcm::new(Key::<Aes256Gcm>::from_slice(key));
    let nonce = Nonce::from_slice(&nonce);
    let plaintext = cipher
        .decrypt(&nonce, ciphertext.as_slice())
        .expect("Failed to decrypt");

    let plaintext_utf8 = String::from_utf8(plaintext).ok()?;

    serde_json::from_str(&plaintext_utf8).ok()
}

#[cfg(test)]
mod tests {
    #[test]
    fn test_wrap_json() {
        let data = "test";
        let key: [u8; 32] = [0; 32];

        let wrapped = super::wrap_json(data, key.as_ref());
        let unwrapped = super::unwrap_json::<String>(&wrapped, key.as_ref()).unwrap();

        assert_eq!(data, unwrapped);
    }
}
