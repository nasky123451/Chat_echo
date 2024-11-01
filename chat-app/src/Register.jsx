import React, { useState, useRef, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Container,
  TextField,
  Button,
  Typography,
  Box,
  Grid,
  Snackbar,
  Alert,
  IconButton,
} from "@mui/material";
import PhoneInput from "react-phone-input-2";
import "react-phone-input-2/lib/style.css";
import RefreshIcon from "@mui/icons-material/Refresh";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import CryptoJS from 'crypto-js';

const Register = () => {
  const navigate = useNavigate();
  const [captcha, setCaptcha] = useState(""); // Captcha 文字
  const [captchaInput, setCaptchaInput] = useState(""); // Captcha 輸入
  const [message, setMessage] = useState("");
  const [openSnackbar, setOpenSnackbar] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false); // 註冊是否成功
  const canvasRef = useRef(null);

  // 隨機生成驗證碼
  const generateCaptcha = () => {
    const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
    let captchaText = "";
    for (let i = 0; i < 6; i++) {
      captchaText += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    setCaptcha(captchaText);

    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.font = "24px Arial";
    ctx.fillText(captchaText, 10, 30);
  };

  useEffect(() => {
    generateCaptcha(); // 初始化生成驗證碼
  }, []);

  // 加密函数
  const encryptData = (data, secretKey) => {
    // 生成随机的 IV
    const iv = CryptoJS.lib.WordArray.random(16);
    
    // 将数据转换为 JSON 字符串
    const jsonString = JSON.stringify(data);
    
    // 使用 AES 加密
    const encryptedData = CryptoJS.AES.encrypt(jsonString, CryptoJS.enc.Utf8.parse(secretKey), {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7,
    });
    
    // 返回加密数据和 IV（可用于解密）
    return {
        encryptedData: encryptedData.toString(),
        iv: iv.toString(),
    };
  };

  const handleRegisterSubmit = async (values) => {
    // 檢查驗證碼是否正確
    if (values.captcha !== captcha) {
      setMessage("Invalid captcha");
      setIsSuccess(false);
      setOpenSnackbar(true);
      return;
    }

    const registerData = {
      username: values.username,
      password: values.password,
      phone: values.phone,
      email: values.email,
    };

    // 使用密鑰加密資料
    const secretKey = "your-secret-key1"; // 請務必妥善管理這個密鑰
    const { encryptedData, iv } = encryptData(registerData, secretKey);

    try {
      const response = await fetch("/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ encryptedData, iv }),
      });

      const data = await response.json();
      if (response.ok) {
        setMessage("User registered successfully");
        setIsSuccess(true); // 設置成功狀態
        setOpenSnackbar(true); // 打開 Snackbar
      } else {
        setMessage(data.error || "Registration failed");
        setIsSuccess(false); // 設置失敗狀態
        setOpenSnackbar(true); // 打開 Snackbar
      }
    } catch (error) {
      console.error("Error:", error); // 打印错误以帮助调试
      setMessage("An error occurred");
      setIsSuccess(false); // 設置失敗狀態
      setOpenSnackbar(true); // 打開 Snackbar
    }
  };

  const handleCloseSnackbar = () => {
    setOpenSnackbar(false);
  };

  // 表單驗證規則
  const validationSchema = Yup.object().shape({
    username: Yup.string().required("Username is required"),
    password: Yup.string().required("Password is required"),
    phone: Yup.string().required("Phone number is required"),
    email: Yup.string().email("Invalid Gmail format").required("Gmail is required"),
    captcha: Yup.string().required("Captcha is required"),
  });

  return (
    <Container maxWidth="xs" sx={{ mt: 5 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Register
      </Typography>
      <Formik
        initialValues={{
          username: "",
          password: "",
          phone: "",
          email: "",
          captcha: "",
        }}
        validationSchema={validationSchema}
        onSubmit={handleRegisterSubmit}
      >
        {({ errors, touched, setFieldValue, values }) => (
          <Form>
            <Box display="flex" flexDirection="column" gap={2}>
              <Field
                as={TextField}
                label="Username"
                name="username"
                variant="outlined"
                fullWidth
                error={touched.username && !!errors.username}
                helperText={touched.username && errors.username}
              />

              <Field
                as={TextField}
                label="Password"
                type="password"
                name="password"
                variant="outlined"
                fullWidth
                error={touched.password && !!errors.password}
                helperText={touched.password && errors.password}
              />

              {/* 手機號選單及格式驗證 */}
              <Field name="phone">
                {({ field, meta }) => (
                  <PhoneInput
                    country={"us"}
                    value={field.value}
                    onChange={(phone) => setFieldValue("phone", phone)}
                    inputStyle={{
                      border: meta.touched && meta.error ? '1px solid red' : '1px solid #ccc',
                      borderRadius: '4px',
                      fontSize: '16px',
                      width: '100%',
                      boxSizing: 'border-box',
                    }}
                    dropdownStyle={{
                      borderRadius: '4px',
                      boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
                      border: '1px solid #ccc',
                    }}
                  />
                )}
              </Field>

              <Field
                as={TextField}
                label="Gmail"
                type="email"
                name="email"
                variant="outlined"
                fullWidth
                error={touched.email && !!errors.email}
                helperText={touched.email && errors.email}
              />

              {/* 驗證碼部分 */}
              <Grid container display="flex" alignItems="center" gap={2}>
                <Grid item xs={6}>
                  <Field
                    as={TextField}
                    label="Captcha"
                    name="captcha"
                    variant="outlined"
                    fullWidth
                    error={touched.captcha && !!errors.captcha}
                    helperText={touched.captcha && errors.captcha}
                  />
                </Grid>
                <Grid item xs={3}>
                  <canvas
                    ref={canvasRef}
                    width="100"
                    height="40"
                    style={{ border: "1px solid #ccc" }}
                  />
                </Grid>
                <Grid item xs={1}>
                  <IconButton onClick={generateCaptcha} aria-label="refresh captcha">
                    <RefreshIcon />
                  </IconButton>
                </Grid>
              </Grid>

              <Button type="submit" variant="contained" color="primary">
                Register
              </Button>
              <Button
                onClick={() => navigate("/")}
                variant="outlined"
                color="secondary"
              >
                Back to Login
              </Button>
            </Box>
          </Form>
        )}
      </Formik>

      <Snackbar open={openSnackbar} autoHideDuration={6000} onClose={handleCloseSnackbar}>
        <Alert onClose={handleCloseSnackbar} severity={isSuccess ? "success" : "error"}>
          {message}
        </Alert>
      </Snackbar>
    </Container>
  );
};

export default Register;
