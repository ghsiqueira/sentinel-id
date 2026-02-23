import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Mail, Lock, ArrowRight, QrCode, Shield, Smartphone, Loader2 } from 'lucide-react';
import api from '../services/api';

export default function Login() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [isPromptMode, setIsPromptMode] = useState(false);
  const [promptRequestId, setPromptRequestId] = useState<string | null>(null);

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await api.post('/auth/login', formData);
      localStorage.setItem('token', response.data.access_token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
      navigate('/dashboard');
    } catch (err: any) {
      setError('Credenciais inválidas. Tente novamente.');
    } finally {
      setLoading(false);
    }
  }

  async function handlePromptInit() {
    if (!formData.email) {
      setError('Digite seu e-mail primeiro para usar a aprovação pelo celular.');
      return;
    }

    setError('');
    setLoading(true);
    try {
      const response = await api.post('/auth/prompt/init', { email: formData.email });
      setPromptRequestId(response.data.request_id);
      setIsPromptMode(true);
    } catch (err: any) {
      setError('Celular de segurança não configurado para este usuário.');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let interval: ReturnType<typeof setInterval>;
    
    if (isPromptMode && promptRequestId) {
      interval = setInterval(async () => {
        try {
          const response = await api.get(`/auth/qr/poll/${promptRequestId}`);
          console.log("RADAR:", response.data);

          const token = response.data.access_token || response.data.token;

          if (token) {
            clearInterval(interval); 
            
            localStorage.setItem('token', token);
            if (response.data.refresh_token) {
              localStorage.setItem('refresh_token', response.data.refresh_token);
            }
            navigate('/dashboard');
            
          } else if (response.data.status?.toUpperCase() === 'REJECTED') {
            clearInterval(interval);
            setError('Login negado pelo dispositivo móvel.');
            setIsPromptMode(false);
            setPromptRequestId(null);
          }
        } catch (err: any) {
          if (err.response && err.response.status === 400) {
             clearInterval(interval);
             setError('Sessão expirada. Tente novamente.');
             setIsPromptMode(false);
             setPromptRequestId(null);
          }
        }
      }, 2000);
    }

    return () => clearInterval(interval);
  }, [isPromptMode, promptRequestId, navigate]);

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-800 p-8 rounded-2xl shadow-2xl w-full max-w-md">
        
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-500/10 mb-4">
            <Shield className="w-8 h-8 text-blue-500" />
          </div>
          <h1 className="text-2xl font-bold text-white">Bem-vindo de volta</h1>
          <p className="text-slate-400 text-sm mt-2">Acesse sua conta Sentinel ID</p>
        </div>

        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm text-center mb-4">
            {error}
          </div>
        )}

        {isPromptMode ? (
          <div className="text-center py-6 animate-in fade-in zoom-in duration-300">
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-emerald-500/10 mb-6 relative">
              <Smartphone className="w-10 h-10 text-emerald-500 relative z-10" />
              <div className="absolute inset-0 border-2 border-emerald-500/30 rounded-full animate-ping"></div>
            </div>
            
            <h2 className="text-xl font-bold text-white mb-2">Verifique seu Celular</h2>
            <p className="text-slate-400 text-sm mb-8 px-4">
              Enviamos um pedido de aprovação para o seu dispositivo principal. Abra o app para liberar o acesso.
            </p>

            <div className="flex justify-center items-center text-emerald-500 gap-2 mb-8">
              <Loader2 className="w-5 h-5 animate-spin" />
              <span className="text-sm font-medium">Aguardando aprovação...</span>
            </div>

            <button 
              onClick={() => { setIsPromptMode(false); setPromptRequestId(null); }}
              className="text-slate-500 hover:text-slate-300 text-sm transition-colors underline"
            >
              Cancelar e usar senha
            </button>
          </div>
        ) : (
          <form onSubmit={handleLogin} className="space-y-4">
            <div>
              <label className="text-xs font-medium text-slate-400 uppercase">E-mail</label>
              <div className="relative mt-1">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 w-5 h-5" />
                <input 
                  type="email" 
                  value={formData.email}
                  onChange={(e) => setFormData({...formData, email: e.target.value})}
                  className="w-full bg-slate-800 border border-slate-700 text-white rounded-lg py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-600"
                  placeholder="seu@email.com"
                  required
                />
              </div>
            </div>

            <div>
              <label className="text-xs font-medium text-slate-400 uppercase">Senha</label>
              <div className="relative mt-1">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 w-5 h-5" />
                <input 
                  type="password" 
                  value={formData.password}
                  onChange={(e) => setFormData({...formData, password: e.target.value})}
                  className="w-full bg-slate-800 border border-slate-700 text-white rounded-lg py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-600"
                  placeholder="••••••••"
                />
              </div>
            </div>

            <button 
              type="submit" 
              disabled={loading}
              className="w-full bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 rounded-lg transition-all flex items-center justify-center gap-2 mt-6 disabled:opacity-50"
            >
              {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : (
                <>Entrar <ArrowRight className="w-5 h-5" /></>
              )}
            </button>

            <div className="relative my-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-800"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-slate-900 text-slate-500">Acesso sem senha</span>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={handlePromptInit}
                disabled={loading || !formData.email}
                className="w-full bg-emerald-600/10 hover:bg-emerald-600/20 text-emerald-500 font-medium py-2.5 rounded-lg transition-all flex items-center justify-center gap-2 border border-emerald-500/20 disabled:opacity-50"
              >
                <Smartphone size={18} />
                Usar Celular
              </button>

              <button
                type="button"
                onClick={() => navigate('/qr-login')}
                className="w-full bg-slate-800 hover:bg-slate-700 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center gap-2 border border-slate-700"
              >
                <QrCode size={18} />
                QR Code
              </button>
            </div>
          </form>
        )}

        {!isPromptMode && (
          <div className="mt-6 text-center">
            <Link to="/register" className="text-slate-400 hover:text-white text-sm transition-colors">
              Não tem uma conta? <span className="text-blue-500">Cadastre-se</span>
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}