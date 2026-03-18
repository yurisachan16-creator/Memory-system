import { createContext, useContext, useState } from "react";
import type { PropsWithChildren } from "react";

interface UserContextValue {
  userId: string;
  draftUserId: string;
  setDraftUserId: (value: string) => void;
  confirmUserId: () => void;
}

const storageKey = "memory-system.active-user-id";

const UserContext = createContext<UserContextValue | null>(null);

function getStoredUserId() {
  if (typeof window === "undefined") {
    return "";
  }
  return window.localStorage.getItem(storageKey)?.trim() || "";
}

export function UserProvider({ children }: PropsWithChildren) {
  const [userId, setUserId] = useState(getStoredUserId);
  const [draftUserId, setDraftUserId] = useState(getStoredUserId);

  const confirmUserId = () => {
    const nextUserId = draftUserId.trim();
    setUserId(nextUserId);

    if (typeof window !== "undefined") {
      if (nextUserId) {
        window.localStorage.setItem(storageKey, nextUserId);
      } else {
        window.localStorage.removeItem(storageKey);
      }
    }
  };

  return (
    <UserContext.Provider
      value={{
        userId,
        draftUserId,
        setDraftUserId,
        confirmUserId
      }}
    >
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const context = useContext(UserContext);
  if (!context) {
    throw new Error("useUser must be used inside UserProvider");
  }
  return context;
}
