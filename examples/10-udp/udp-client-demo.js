// UDP 客户端示例
const { net } = require('net');

console.log('=== UDP Client Example ===\n');

// 创建 UDP 套接字
const socket = net.createUDPSocket('udp4');

// 可选：绑定本地端口以接收回复
socket.bind('0', '0.0.0.0', () => {
    console.log('UDP Client ready');
}).then(() => {
    const addr = socket.address();
    console.log('Bound to', addr.address + ':' + addr.port);
    console.log();
    
    // 接收回复
    socket.on('message', (msg, rinfo) => {
        console.log('Received reply from', rinfo.address + ':' + rinfo.port);
        console.log('Reply:', msg.trim());
        console.log();
    });
    
    // 发送消息到服务器
    console.log('Sending messages to UDP server...\n');
    
    const messages = [
        'Hello UDP Server!',
        'This is message 1',
        'This is message 2',
        'Testing UDP communication',
        'Goodbye!'
    ];
    
    let delay = 0;
    messages.forEach((msg, index) => {
        setTimeout(() => {
            console.log('Sending:', msg);
            socket.send(msg + '\n', '9090', 'localhost')
                .then(() => {
                    console.log('Message sent successfully');
                })
                .catch(err => {
                    console.error('Failed to send message:', err.message);
                });
        }, delay);
        delay += 1000;  // 每秒发送一条消息
    });
    
    // 5 秒后关闭套接字
    setTimeout(() => {
        console.log('\nClosing socket...');
        socket.close();
    }, 6000);
    
}).catch(err => {
    console.error('Failed to bind socket:', err.message);
});

// 不绑定本地端口的简单发送示例
console.log('Simple send example (no reply expected):');
const simpleSocket = net.createUDPSocket('udp4');

simpleSocket.send('Quick message\n', '9090', 'localhost', () => {
    console.log('Quick message sent!');
}).then(() => {
    console.log('Send operation completed\n');
}).catch(err => {
    console.error('Send failed:', err.message);
});
