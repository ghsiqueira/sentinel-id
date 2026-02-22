import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Mail, Lock, ArrowRight, QrCode, Shield } from 'lucide-react';
import api from '../services/api';

export default function Login() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

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

        <form onSubmit={handleLogin} className="space-y-4">
          
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm text-center">
              {error}
            </div>
          )}

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
                required
              />
            </div>
          </div>

          <button 
            type="submit" 
            disabled={loading}
            className="w-full bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 rounded-lg transition-all flex items-center justify-center gap-2 mt-6 disabled:opacity-50"
          >
            {loading ? 'Entrando...' : (
              <>
                Entrar <ArrowRight className="w-5 h-5" />
              </>
            )}
          </button>

          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-slate-800"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-slate-900 text-slate-500">Ou entre com</span>
            </div>
          </div>

          <button
            type="button"
            onClick={() => navigate('/qr-login')}
            className="w-full bg-slate-800 hover:bg-slate-700 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center gap-2 border border-slate-700"
          >
            <QrCode size={18} />
            Escanear QR Code
          </button>
        </form>

        <div className="mt-6 text-center">
          <Link to="/register" className="text-slate-400 hover:text-white text-sm transition-colors">
            Não tem uma conta? <span className="text-blue-500">Cadastre-se</span>
          </Link>
        </div>

      </div>
    </div>
  );
}