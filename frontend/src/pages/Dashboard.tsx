import { useNavigate } from 'react-router-dom';
import { LogOut, User, ShieldAlert, CheckCircle } from 'lucide-react';
import axios from 'axios';
import { useEffect, useState } from 'react';

export function Dashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<{ message: string; your_id: string } | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const token = localStorage.getItem('access_token');
        if (!token) {
            navigate('/login');
            return;
        }

        const response = await axios.get('http://localhost:8080/api/users/me', {
          headers: { Authorization: `Bearer ${token}` }
        });

        setUser(response.data);
      } catch (error) {
        console.error("Erro ao buscar usuário:", error);
        navigate('/login');
      }
    };

    fetchUser();
  }, [navigate]);

  function handleLogout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    navigate('/login');
  }

  async function handleKillSwitch() {
    const confirm = window.confirm("Tem certeza? Isso vai desconectar você de TODOS os celulares e computadores.");
    if (!confirm) return;

    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      
      await axios.post('http://localhost:8080/api/users/revoke-all', {}, {
        headers: { Authorization: `Bearer ${token}` }
      });

      alert("SUCESSO: Todas as outras sessões foram derrubadas!");
      handleLogout(); 

    } catch (error) {
      alert("Erro ao revogar sessões.");
      console.error(error);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 text-white p-8">
      <div className="max-w-4xl mx-auto">
        
        {/* Cabeçalho */}
        <div className="flex items-center justify-between mb-12">
          <h1 className="text-3xl font-bold text-blue-500 tracking-tight">Sentinel<span className="text-white">Dashboard</span></h1>
          <button 
            onClick={handleLogout}
            className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors"
          >
            <LogOut size={20} />
            Sair
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            
            <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-xl">
              <div className="flex items-center gap-4 mb-6">
                <div className="bg-blue-500/10 p-3 rounded-lg">
                  <User size={24} className="text-blue-500" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-white">Sua Identidade</h2>
                  <p className="text-slate-400 text-sm">Dados da sessão segura</p>
                </div>
              </div>

              {user ? (
                <div className="space-y-4">
                  <div className="flex items-center gap-2 text-green-400 bg-green-400/10 p-3 rounded-lg border border-green-400/20">
                    <CheckCircle size={18} />
                    <span className="text-sm font-medium">{user.message}</span>
                  </div>
                  <div className="p-3 bg-slate-950 rounded-lg border border-slate-800">
                    <p className="text-xs text-slate-500 uppercase mb-1">User ID</p>
                    <p className="font-mono text-slate-300 text-sm break-all">{user.your_id}</p>
                  </div>
                </div>
              ) : (
                <div className="animate-pulse flex space-x-4">
                    <div className="h-12 bg-slate-800 rounded w-full"></div>
                </div>
              )}
            </div>

            <div className="bg-slate-900 border border-red-900/30 rounded-xl p-6 shadow-xl relative overflow-hidden group">
                <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                    <ShieldAlert size={100} className="text-red-500" />
                </div>

                <div className="flex items-center gap-4 mb-6 relative z-10">
                    <div className="bg-red-500/10 p-3 rounded-lg">
                    <ShieldAlert size={24} className="text-red-500" />
                    </div>
                    <div>
                    <h2 className="text-lg font-semibold text-white">Zona de Perigo</h2>
                    <p className="text-red-400/80 text-sm">Ações irreversíveis</p>
                    </div>
                </div>

                <p className="text-slate-400 text-sm mb-6 relative z-10">
                    Se você suspeita que sua conta foi comprometida, use este botão para desconectar <strong>todos</strong> os dispositivos imediatamente.
                </p>

                <button 
                    onClick={handleKillSwitch}
                    disabled={loading}
                    className="relative z-10 w-full bg-red-600 hover:bg-red-500 text-white font-bold py-3 rounded-lg transition-all shadow-lg hover:shadow-red-900/20 flex items-center justify-center gap-2"
                >
                    {loading ? 'Processando...' : '☢️ REVOGAR TODAS AS SESSÕES'}
                </button>
            </div>

        </div>
      </div>
    </div>
  );
}