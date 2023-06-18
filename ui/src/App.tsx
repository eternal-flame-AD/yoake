
import './App.css'
import { useState } from 'react';
import { RouterProvider, createHashRouter } from 'react-router-dom';
import PageBase from './PageBase';
import { ThemeProvider } from '@emotion/react';
import theme from './theme';
import { LoginContext } from './context/LoginContext';
import { EmptyAuthInfo, getLoginState } from './api/auth';
import HomePage from './pages/HomePage';
import LoginDialog from './pages/LoginDialog';
import GradesPage from './pages/GradesPage';
import FramePage from './pages/FramePage';
import { EnsureRole } from './components/EnsureRole';
import MedsPage from './pages/MedsPage';


const persistent_pages = [
  {
    path: "/gotify_ui",
    element: <FramePage url="https://gotify.yumechi.jp/" />
  }
];

function App() {

  const [auth, setAuth] = useState(EmptyAuthInfo);
  const [mounted, setMounted] = useState(false);

  const refreshAuth = () => {
    getLoginState().then((auth) => {
      setAuth(auth);
    }
    ).catch((error) => {
      console.error(error);
      setAuth(EmptyAuthInfo);
    });
  }

  if (!mounted) {
    setMounted(true);
    refreshAuth();
  }

  const router = createHashRouter([
    {
      path: "/",
      element: <PageBase persistentPages={persistent_pages} ><HomePage /></PageBase>
    },
    {
      path: "/grades",
      element: <PageBase persistentPages={persistent_pages} ><EnsureRole role="Admin"><GradesPage /></EnsureRole></PageBase>
    },
    {
      path: "/meds",
      element: <PageBase persistentPages={persistent_pages} ><EnsureRole role="Admin"><MedsPage /></EnsureRole></PageBase>
    },
    {
      path: "/gotify_ui",
      element: <PageBase persistentPages={persistent_pages} ></PageBase>
    },
    {
      path: "/login",
      element: <PageBase persistentPages={persistent_pages} ><LoginDialog open={true}
        onClose={(_, navigate) => navigate("/")}
      /></PageBase>
    }
  ]);

  return (
    <ThemeProvider theme={theme}>
      <LoginContext.Provider value={{
        auth,
        setAuth: setAuth,
        refreshAuth: refreshAuth,
      }}>
        <RouterProvider router={router} />
      </LoginContext.Provider>
    </ThemeProvider>
  )
}

export default App
