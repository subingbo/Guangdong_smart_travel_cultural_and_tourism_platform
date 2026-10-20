import { createContext, useCallback, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { api, AuthResponse, getToken, setToken, setLanguage, User } from '../api/client';

interface AuthState {
  user: User | null;
  signIn: (username: string, password: string) => Promise<void>;
  signUp: (username: string, password: string, nickname?: string) => Promise<void>;
  signOut: () => void;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthState>(null as unknown as AuthState);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const { i18n } = useTranslation();

  const refresh = useCallback(async () => {
    if (!getToken()) {
      setUser(null);
      return;
    }
    try {
      setUser(await api.get<User>('/api/me'));
    } catch {
      setToken(null);
      setUser(null);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const applyAuth = (r: AuthResponse) => {
    setToken(r.token);
    setUser(r.user);
    if (r.user.language) {
      setLanguage(r.user.language);
      void i18n.changeLanguage(r.user.language);
    }
  };

  const signIn = async (username: string, password: string) =>
    applyAuth(await api.post<AuthResponse>('/api/auth/login', { username, password }));

  const signUp = async (username: string, password: string, nickname?: string) =>
    applyAuth(await api.post<AuthResponse>('/api/auth/register', { username, password, nickname }));

  const signOut = () => {
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, signIn, signUp, signOut, refresh }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
