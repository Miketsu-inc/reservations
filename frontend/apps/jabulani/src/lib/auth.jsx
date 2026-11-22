import { useRouter } from "@tanstack/react-router";
import { createContext, useCallback, useContext } from "react";
import { hasPermission } from "./permissions";

const AuthContext = createContext();

export function AuthProvider({ children, authContext }) {
  const router = useRouter();

  const switchMerchant = useCallback(
    (merchantId) => {
      localStorage.setItem("activeMerchantId", merchantId);
      router.invalidate();
    },
    [router]
  );

  return (
    <AuthContext.Provider
      value={{
        ...authContext,
        switchMerchant,
        can: (permission) =>
          hasPermission(authContext?.permissions, permission),
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
