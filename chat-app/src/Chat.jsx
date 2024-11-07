import React, { useState, useEffect, useRef } from 'react';
import { jwtDecode } from 'jwt-decode';
import {
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemText,
  Typography,
  TextField,
  Button,
  Box,
  Divider,
  Card,
} from '@mui/material';
import { Menu as MenuIcon, Close as CloseIcon, ArrowDownward as ArrowDownwardIcon } from '@mui/icons-material';
import { Tooltip } from '@mui/material';

const Chat = () => {
  const [currentUser, setCurrentUser] = useState('');
  const [onlineUsers, setOnlineUsers] = useState([]);
  const [offlineUsers, setOfflineUsers] = useState([]);
  const [messages, setMessages] = useState([]);
  const [messageInput, setMessageInput] = useState('');
  const [ws, setWs] = useState(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const messageEndRef = useRef(null);
  const chatContainerRef = useRef(null);
  const [noMoreMessages, setNoMoreMessages] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [isAutoScroll, setIsAutoScroll] = useState(true);
  const [showArrow, setShowArrow] = useState(true);
  const autoScrollIntervalRef = useRef(null);

  const connectWebSocket = () => {
    const token = localStorage.getItem('token');
    const ws = new WebSocket('ws://localhost:8080/ws'); // 替換為你的 WebSocket URL

    ws.onopen = () => {
      console.log('WebSocket 連線已開啟');
      ws.send(JSON.stringify({ type: "auth", token }));
      setWs(ws);
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
    
        if (msg.type === "message") {
          setMessages((prevMessages) => [...prevMessages, msg]);
          if (isAutoScroll) {
            scrollToBottom();
          } else {
            setShowArrow(true);
          }
          if (chatContainerRef.current.scrollHeight - chatContainerRef.current.scrollTop === chatContainerRef.current.clientHeight) {
            scrollToBottom();
          }
        } else if (msg.type === "userStatus") {
          updateUserStatus(msg.username, msg.status);
        }
      } catch (error) {
        console.error('Error handling message:', error);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket 連線已關閉，嘗試重新連線...');
      setIsConnected(false);
      // 嘗試在 2 秒後重新連線
      setTimeout(connectWebSocket, 2000);
    };

    ws.onerror = (error) => {
      console.error('WebSocket 發生錯誤', error);
    };

    setWs(ws);
  };
  
  // 初始数据加载
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      window.location.href = '/';
    }

    connectWebSocket();

    const loadInitialData = async () => {
      if (token) {
        try {
          const decoded = jwtDecode(token);
          setCurrentUser(decoded.username);
          await fetchOnlineUsers();
  
          const latestDate = await fetchLatestChatDate('general');
          if (!latestDate) {
            setNoMoreMessages(true);
          }
        } catch (error) {
          console.error("Token decoding error:", error);
        }
      }
    };
  
    loadInitialData();
  
    return () => {
      if (ws) {
        ws.close();
        console.log('WebSocket 連線已關閉');
      }
    };
  }, []); // 空依赖数组，确保只在组件首次加载时执行

  // 获取在线用户
  const fetchOnlineUsers = async () => {
    try {
      const response = await fetch('/online-users', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      const data = await response.json();
      setOnlineUsers(data.onlineUsers || []);
      setOfflineUsers(data.offlineUsers || []);
    } catch (error) {
      console.error('Failed to fetch online users:', error);
    }
  };

  // 获取最新的聊天日期
  const fetchLatestChatDate = async (room) => {
    if (noMoreMessages) {
      return null;
    }
    if (!room) {
      console.error('Room parameter is required');
      return null;
    }

    try {
      const response = await fetch(`/latest-chat-date?room=${encodeURIComponent(room)}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      if (Array.isArray(data.totalMessages)) {
        if (data.totalMessages.length === 0) {
          console.warn('No messages found for the selected date.');
          setNoMoreMessages(true);
        } else {
          setNoMoreMessages(false);
          setMessages((prevMessages) => [...data.totalMessages, ...prevMessages]);
          setTimeout(() => {
            chatContainerRef.current.scrollTop += messageEndRef.current.clientHeight;
          }, 0);
        }
      } else {
        console.error('Messages is not an array or is undefined:', data.totalMessages);
      }

      if (data.message === "沒有更多資料") {
        return null;
      }

      return data.latestChatDate;
    } catch (error) {
      console.error('Failed to fetch latest chat date:', error);
      return null;
    }
  };



  // 滚动到聊天底部
  const scrollToBottom = () => {
    if (messageEndRef.current) {
      messageEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  };

  // 处理聊天记录获取
  const fetchMessages = async (date, room, separator) => {
    if (noMoreMessages) {
      return null;
    }
    
    try {
      const response = await fetch(`/chat-history?date=${date || new Date().toISOString().split('T')[0]}&room=${room}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      if (Array.isArray(data.messages)) {
        if (data.messages.length === 0) {
          console.warn('No messages found for the selected date.');
          setNoMoreMessages(true);
        } else {
          setNoMoreMessages(false);

          if (separator) {
            const previousDate = new Date();
            previousDate.setDate(previousDate.getDate() - 1);

            const dateSeparatorMessage = {
              isSeparator: true,
              date: previousDate.toISOString(),
            };

            setMessages((prevMessages) => [
              ...data.messages,
              dateSeparatorMessage,
              ...prevMessages,
            ]);
          } else {
            setMessages((prevMessages) => [...data.messages, ...prevMessages]);
          }

          setTimeout(() => {
            chatContainerRef.current.scrollTop += messageEndRef.current.clientHeight;
          }, 0);
        }
      } else {
        console.error('Messages is not an array or is undefined:', data.messages);
      }
    } catch (error) {
      console.error('Failed to fetch messages:', error);
    }
  };

  // 处理用户状态更新
  const updateUserStatus = (username, status) => {
    if (status === 'online') {
      setOnlineUsers((prev) => [...new Set([...prev, username])]);
      setOfflineUsers((prev) => prev.filter(user => user !== username));
    } else if (status === 'offline') {
      setOfflineUsers((prev) => [...new Set([...prev, username])]);
      setOnlineUsers((prev) => prev.filter(user => user !== username));
    }
  };

  // 处理消息发送
  const sendMessage = (e) => {
    e.preventDefault();
    if (!messageInput || !ws) return;

    const message = {
      type: "message",
      room: 'general',
      sender: currentUser,
      content: messageInput,
      time: new Date().toISOString(),
    };

    ws.send(JSON.stringify(message));
    setMessageInput('');
  };

  const handleScroll = (e) => {
    const { scrollTop } = e.target;

    if (scrollTop === 0) {
      const currentMessagesCount = messages.length;

      const lastMessageDate = messages.length > 0 ? new Date(messages[1].time) : null;
      const previousDate = lastMessageDate
        ? new Date(lastMessageDate.setDate(lastMessageDate.getDate() - 1)).toISOString().split('T')[0]
        : null;

      fetchMessages(previousDate, 'general', true);
    }

    setIsAutoScroll(scrollTop + chatContainerRef.current.clientHeight >= chatContainerRef.current.scrollHeight);
    setShowArrow(!isAutoScroll);
  };

  const handleMouseWheel = (e) => {
    setIsAutoScroll(e.deltaY >= 0); // 滾動向下
  };

  const handleArrowClick = () => {
    setIsAutoScroll(true);
    scrollToBottom();
    autoScrollIntervalRef.current = setInterval(scrollToBottom, 1000); // 每秒滾動到底部
  };

  useEffect(() => {
    const handleMouseWheelScroll = (event) => {
      if (event.deltaY < 0) {
        clearInterval(autoScrollIntervalRef.current); // 停止自動滾動
      }
    };

    window.addEventListener('wheel', handleMouseWheelScroll);

    return () => {
      clearInterval(autoScrollIntervalRef.current);
      window.removeEventListener('wheel', handleMouseWheelScroll);
    };
  }, []);

  // 登出功能
  const logout = () => {
    if (ws) {
      ws.send(JSON.stringify({ type: "logout", username: currentUser })); // 发送登出消息
      ws.close(); // 关闭 WebSocket 连接
      console.log('User logged out');
    }
    // 清理本地存储和其他用户状态
    localStorage.removeItem('token');
    setCurrentUser('');
    setMessages([]);
    setOnlineUsers([]);
    setOfflineUsers([]);
    window.location.href = '/';
  };

  return (
    <div style={{ display: 'flex', height: '100vh' }}>
      {/* Drawer components */}
      <Drawer
        anchor="left"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        variant="temporary"
      >
        <Box sx={{ width: 250, padding: '20px', backgroundColor: '#f0f0f0' }}>
          <IconButton onClick={() => setDrawerOpen(false)} sx={{ marginBottom: '16px' }}>
            <CloseIcon />
          </IconButton>
          <Typography variant="h6" gutterBottom>當前用戶</Typography>
          <Typography variant="body1" sx={{ marginBottom: '16px' }}>{currentUser}</Typography>
          <Divider sx={{ marginBottom: '16px' }} />
          <Typography variant="h6" gutterBottom>在線用戶</Typography>
          <List>
            {(onlineUsers || [])
              .filter(user => user !== currentUser)
              .map((user, index) => (
                <ListItem key={index} sx={{ '&:hover': { backgroundColor: '#e0e0e0' } }}>
                  <ListItemText primary={user} />
                </ListItem>
              ))}
          </List>
          <Divider sx={{ marginBottom: '16px' }} />
          <Typography variant="h6" gutterBottom>離線用戶</Typography>
          <List>
            {(offlineUsers || [])
              .filter(user => user !== currentUser)
              .map((user, index) => (
                <ListItem key={index} sx={{ '&:hover': { backgroundColor: '#e0e0e0' } }}>
                  <ListItemText primary={user} />
                </ListItem>
              ))}
          </List>
        </Box>
      </Drawer>
  
      {/* Main content */}
      <div style={{ flexGrow: 1, padding: '20px', display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <IconButton onClick={() => setDrawerOpen(true)}>
          <MenuIcon />
        </IconButton>

        <Typography variant="h4" align="center" sx={{ flexGrow: 1 }}>
          聊天窗口
        </Typography>

        <Button 
          variant="contained" 
          color="secondary" 
          onClick={logout} 
        >
          登出
        </Button>
      </Box>
  
        <Card sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', position: 'relative', mb: 2, padding: 2 }}>
          <Box 
            ref={chatContainerRef}
            sx={{ 
              flexGrow: 1,  // 让 Box 占满剩余空间
              overflowY: 'auto',  // 确保超出时可以滚动
              border: '1px solid #ccc', 
              padding: '10px', 
              backgroundColor: '#f9f9f9' 
            }}
            onScroll={handleScroll}
            onWheel={handleMouseWheel}
          >
            {noMoreMessages && (
              <Typography variant="body2" color="textSecondary" align="center" style={{ marginBottom: '10px' }}>
                没有更多消息了
              </Typography>
            )}
            {(messages || []).map((msg, index) => (
              <div key={index}>
                {msg.isSeparator ? (
                  <Typography variant="body2" align="center" sx={{ margin: '10px 0', fontWeight: 'bold' }}>
                    {new Date(msg.date).toLocaleDateString()} {/* Show date */}
                  </Typography>
                ) : (
                  <Box sx={{ marginBottom: 2, padding: 1, border: '1px solid #e0e0e0', borderRadius: 2, backgroundColor: '#f9f9f9' }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <strong style={{ color: '#3f51b5' }}>{msg.sender}:</strong>
                      <em style={{ fontSize: '0.8em', color: '#888' }}>
                        {new Date(msg.time).toLocaleString('zh-CN', {
                          year: 'numeric',
                          month: '2-digit',
                          day: '2-digit',
                          hour: '2-digit',
                          minute: '2-digit',
                          second: '2-digit',
                          hour12: false,
                        })}
                      </em>
                    </Box>
                    <Box sx={{ marginTop: 0.5 }}>
                      {msg.content}
                    </Box>
                  </Box>
                )}
              </div>
            ))}
            <div ref={messageEndRef} />
          </Box>
          <Tooltip title="自動至底" arrow>
            {showArrow && (
              <IconButton 
                onClick={handleArrowClick}
                sx={{ 
                  position: 'absolute', 
                  bottom: 16,
                  right: 16,
                  zIndex: 1
                }}
              >
                <ArrowDownwardIcon />
              </IconButton>
            )}
          </Tooltip>
        </Card>
        <form onSubmit={sendMessage}>
          <TextField 
            value={messageInput}
            onChange={(e) => setMessageInput(e.target.value)}
            label="輸入消息..."
            fullWidth
            variant="outlined"
            sx={{ mb: 1 }}
          />
          <Button type="submit" variant="contained" color="primary">發送</Button>
        </form>
      </div>
    </div>
  );
};

export default Chat;
