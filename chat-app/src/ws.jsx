import React, { useEffect, useState } from 'react';

function WebSocketComponent() {
  const [message, setMessage] = useState('');
  const [serverMessage, setServerMessage] = useState('');
  const [ws, setWs] = useState(null);
  useEffect(() => {
    // 建立 WebSocket 連接
    const token = localStorage.getItem('token');
    const socket = new WebSocket(`ws://localhost:8080/ws`);
    // 當 WebSocket 連接成功時觸發
    socket.onopen = () => {
        console.log('WebSocket connection established');
        socket.send(JSON.stringify({ type: "auth", token }));
        setWs(socket)
    };

    // 當從服務器收到消息時觸發
    socket.onmessage = (event) => {
      console.log('收到服務器消息:', event.data);
      setServerMessage(event.data); // 更新收到的消息
    };

    // 當 WebSocket 連接關閉時觸發
    socket.onclose = () => {
      console.log('WebSocket 連接關閉');
    };

    // 當 WebSocket 出現錯誤時觸發
    socket.onerror = (error) => {
      console.error('WebSocket 錯誤:', error);
    };

    // 在組件卸載時關閉 WebSocket 連接
    return () => {
        socket.close();
    };
  }, []);
  // 發送消息到服務器
  const sendMessage = () => {
    console.log(message);
    ws.send(message);
  };

  return (
    <div>
      <h1>WebSocket 測試</h1>
      <input
        type="text"
        value={message}
        onChange={(e) => setMessage(e.target.value)}></input>
      <button onClick={sendMessage}>發送消息</button>
      <p>伺服器返回的消息：{serverMessage}</p>
    </div>
  );
}

export default WebSocketComponent;