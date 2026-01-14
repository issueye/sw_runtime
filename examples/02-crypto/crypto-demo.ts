// crypto-demo.ts - 加密功能演示

const { crypto } = require('utils');

console.log('=== 加密功能演示 ===');

// 1. 哈希函数演示
console.log('\n1. 哈希函数:');
const data = 'Hello, Crypto World!';
console.log('原始数据:', data);
console.log('MD5:', crypto.md5(data));
console.log('SHA1:', crypto.sha1(data));
console.log('SHA256:', crypto.sha256(data));
console.log('SHA512:', crypto.sha512(data));

// 2. Base64 编解码
console.log('\n2. Base64 编解码:');
const text = 'Hello, Base64!';
const encoded = crypto.base64Encode(text);
const decoded = crypto.base64Decode(encoded);
console.log('原始文本:', text);
console.log('Base64 编码:', encoded);
console.log('Base64 解码:', decoded);
console.log('匹配:', text === decoded);

// 3. Hex 编解码
console.log('\n3. Hex 编解码:');
const hexEncoded = crypto.hexEncode(text);
const hexDecoded = crypto.hexDecode(hexEncoded);
console.log('原始文本:', text);
console.log('Hex 编码:', hexEncoded);
console.log('Hex 解码:', hexDecoded);
console.log('匹配:', text === hexDecoded);

// 4. AES 加解密
console.log('\n4. AES 加解密:');
const secret = 'This is a very secret message!';
const key = 'my-super-secret-key-32-bytes!!';
console.log('原始消息:', secret);
console.log('密钥:', key);

const encrypted = crypto.aesEncrypt(secret, key);
console.log('加密后:', encrypted.substring(0, 50) + '...');

const decrypted = crypto.aesDecrypt(encrypted, key);
console.log('解密后:', decrypted);
console.log('匹配:', secret === decrypted);

// 5. 随机数生成
console.log('\n5. 随机数生成:');
console.log('8字节随机数:', crypto.randomBytes(8));
console.log('16字节随机数:', crypto.randomBytes(16));
console.log('32字节随机数:', crypto.randomBytes(32));

export { crypto };

// 如果在 CommonJS 环境中运行
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { crypto };
}