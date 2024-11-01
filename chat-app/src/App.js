import React from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import Login from "./Login";
import Register from "./Register";
import Chat from './Chat';
import WebSocketComponent from './ws';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/chat" element={<Chat />} />
        <Route path="/WebSocketComponent" element={<WebSocketComponent />} />
      </Routes>
    </Router>
  );
}

export default App;