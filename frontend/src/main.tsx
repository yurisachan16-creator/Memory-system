import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { App as AntdApp, ConfigProvider } from "antd";

import App from "./App";
import { UserProvider } from "./context/UserContext";
import "./styles.css";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: "#be5b3d",
          colorInfo: "#2d6472",
          colorSuccess: "#49755c",
          colorWarning: "#d7903b",
          colorError: "#b44343",
          borderRadius: 18,
          fontFamily: "\"Trebuchet MS\", \"Lucida Sans Unicode\", sans-serif"
        }
      }}
    >
      <AntdApp>
        <BrowserRouter>
          <UserProvider>
            <App />
          </UserProvider>
        </BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  </React.StrictMode>
);
