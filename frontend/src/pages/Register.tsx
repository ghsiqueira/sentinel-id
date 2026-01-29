import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { UserPlus, Mail, Lock, User, FileText, ArrowRight, CheckCircle } from 'lucide-react';
import axios from 'axios';

export function Register() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    full_name: '',
    email: '',
    cpf: '',
    password: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await axios.post('http://localhost:8080/api/auth/register', formData);
      
      alert("Conta criada com sucesso! Faça login.");
      navigate('/login');
      
    } catch (err: any) {
      console.error(err);
      const msg = err.response?.data?.error || 'Erro ao criar conta. Verifique os dados.';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-800 p-8 rounded-2xl shadow-2xl w-full max-w-md">
        
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-500/10 mb-4">
            <UserPlus className="w-8 h-8 text-blue-500" />
          </div>
          <h1 className="text-2xl font-bold text-white">Criar Nova Conta</h1>
          <p className="text-slate-400 text-sm mt-2">Junte-se ao Sentinel ID</p>
        </div>

        <form onSubmit={handleRegister} className="space-y-4">
          
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm text-center">
              {error}
            </div>
          )}

          <div>
            <label className="text-xs font-medium text-slate-400 uppercase">Nome Completo</label>
            <div className="relative mt-1">
              <User className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 w-5 h-5" />
              <input 
                type="text" 
                value={formData.full_name}
                onChange={(e) => setFormData({...formData, full_name: e.target.value})}
                className="w-full bg-slate-800 border border-slate-700 text-white rounded-lg py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-600"
                placeholder="Ex: João da Silva"
                required
              />
            </div>
          </div>

          <div>
            <label className="text-xs font-medium text-slate-400 uppercase">CPF</label>
            <div className="relative mt-1">
              <FileText className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 w-5 h-5" />
              <input 
                type="text" 
                value={formData.cpf}
                onChange={(e) => setFormData({...formData, cpf: e.target.value})}
                className="w-full bg-slate-800 border border-slate-700 text-white rounded-lg py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-600"
                placeholder="12345678900"
                required
              />
            </div>
          </div>

          <div>
            <label className="text-xs font-medium text-slate-400 uppercase">E-mail</label>
            <div className="relative mt-1">
              <Mail className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 w-5 h-5" />
              <input 
                type="email" 
                value={formData.email}
                onChange={(e) => setFormData({...formData, email: e.target.value})}
                className="w-full bg-slate-800 border border-slate-700 text-white rounded-lg py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-600"
                placeholder="joao@email.com"
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
            className="w-full bg-green-600 hover:bg-green-500 text-white font-semibold py-3 rounded-lg transition-all flex items-center justify-center gap-2 mt-6 disabled:opacity-50"
          >
            {loading ? 'Cadastrando...' : (
              <>
                Criar Conta <CheckCircle className="w-5 h-5" />
              </>
            )}
          </button>
        </form>

        <div className="mt-6 text-center">
          <Link to="/login" className="text-slate-400 hover:text-white text-sm transition-colors flex items-center justify-center gap-2">
             Já tem uma conta? <span className="text-blue-500">Faça Login</span>
          </Link>
        </div>

      </div>
    </div>
  );
}