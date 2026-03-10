import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { DashboardPage } from "@/pages/DashboardPage"
import { ClusterCreatePage } from "@/pages/ClusterCreatePage"
import { ClusterDetailPage } from "@/pages/ClusterDetailPage"
import { LoginPage } from "@/pages/LoginPage"
import { AuthProvider, useAuth } from "@/hooks/useAuth"
import { useTheme } from "@/hooks/useTheme"
import { changeLanguage } from "@/i18n"
import { Button } from "@/components/ui/button"
import { Toaster } from "@/components/ui/sonner"
import "@/i18n" // initialize i18next

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <>{children}</>
}

function AppLayout({ children }: { children: React.ReactNode }) {
  const { logout } = useAuth()
  const { isDark, toggle } = useTheme()
  const { t, i18n } = useTranslation()
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto flex h-14 items-center justify-between px-4">
          <a href="/" className="flex items-center gap-2 text-lg font-bold">
            <img src="/icon.png" alt="DST DS Panel" className="h-8 w-8 rounded" />
            {t("app.title")}
          </a>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => changeLanguage(i18n.language === "en" ? "zh" : "en")}
            >
              {i18n.language === "en" ? "中文" : "EN"}
            </Button>
            <Button variant="ghost" size="sm" onClick={toggle}>
              {isDark ? t("app.light") : t("app.dark")}
            </Button>
            <Button variant="ghost" size="sm" onClick={logout}>
              {t("app.logout")}
            </Button>
          </div>
        </div>
      </header>
      <main className="container mx-auto px-4 py-6">{children}</main>
    </div>
  )
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/*"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <Routes>
                    <Route path="/" element={<DashboardPage />} />
                    <Route path="/clusters/new" element={<ClusterCreatePage />} />
                    <Route path="/clusters/:id" element={<ClusterDetailPage />} />
                  </Routes>
                </AppLayout>
              </ProtectedRoute>
            }
          />
        </Routes>
      </BrowserRouter>
      <Toaster richColors position="top-right" />
    </AuthProvider>
  )
}

export default App
