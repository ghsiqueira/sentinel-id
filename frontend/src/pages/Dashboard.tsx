import { useNavigate } from 'react-router-dom';
import { LogOut, User, ShieldAlert, CheckCircle, Monitor, Smartphone, Clock, XCircle } from 'lucide-react';
import axios from 'axios';
import { useEffect, useState } from 'react';

interface Session {
  id: string;
  device_name: string;
  ip_address: string;
  is_revoked: boolean;
  expires_at: string;
}

export function Dashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<{ message: string; your_id: string } | null>(null);
  const [sessions, setSessions] = useState<Session[]>([]); 
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const token = localStorage.getItem('access_token');
        if (!token) {
            navigate('/login');
            return;
        }

        const config = { headers: { Authorization: `Bearer ${token}` } };

        const userRes = await axios.get('http://localhost:8080/api/users/me', config);
        setUser(userRes.data);

        const sessionsRes = await axios.get('http://localhost:8080/api/users/sessions', config);
        setSessions(sessionsRes.data);

      } catch (error) {
        console.error("Erro ao carregar dashboard:", error);
        navigate('/login');
      }
    };

    fetchData();
  }, [navigate]);

  function handleLogout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    navigate('/login');
  }

  async function handleKillSwitch() {
    const confirm = window.confirm("Tem certeza? Isso vai desconectar você de TODOS os lugares.");
    if (!confirm) return;

    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      await axios.post('http://localhost:8080/api/users/revoke-all', {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert("SUCESSO: Todas as sessões foram derrubadas!");
      handleLogout();
    } catch (error) {
      alert("Erro ao revogar sessões.");
    } finally {
      setLoading(false);
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('pt-BR');
  };

  return (
    <div className="min-h-screen bg-slate-950 text-white p-4 md:p-8">
      <div className="max-w-5xl mx-auto space-y-8">
        
        {/* Cabeçalho */}
        <div className="flex items-center justify-between">
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
              <div className="flex items-center gap-4 mb-4">
                <div className="bg-blue-500/10 p-3 rounded-lg">
                  <User size={24} className="text-blue-500" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-white">Minha Identidade</h2>
                  <p className="text-slate-400 text-sm">Status da Conexão</p>
                </div>
              </div>
              {user && (
                <div className="p-3 bg-slate-950 rounded-lg border border-slate-800 font-mono text-sm text-green-400 flex gap-2 items-center">
                   <CheckCircle size={16}/> {user.message}
                </div>
              )}
            </div>

            <div className="bg-slate-900 border border-red-900/30 rounded-xl p-6 shadow-xl relative overflow-hidden group">
                <div className="flex items-center gap-4 mb-4 relative z-10">
                    <div className="bg-red-500/10 p-3 rounded-lg">
                        <ShieldAlert size={24} className="text-red-500" />
                    </div>
                    <div>
                        <h2 className="text-lg font-semibold text-white">Zona de Perigo</h2>
                        <p className="text-red-400/80 text-sm">Kill Switch Global</p>
                    </div>
                </div>
                <button 
                    onClick={handleKillSwitch}
                    disabled={loading}
                    className="relative z-10 w-full bg-red-600 hover:bg-red-500 text-white font-bold py-3 rounded-lg transition-all shadow-lg flex items-center justify-center gap-2"
                >
                    {loading ? 'Processando...' : '☢️ DESCONECTAR TUDO'}
                </button>
            </div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-xl">
            <div className="p-6 border-b border-slate-800">
                <h2 className="text-xl font-semibold flex items-center gap-2">
                    <Monitor className="text-blue-500" /> 
                    Dispositivos Conectados
                </h2>
                <p className="text-slate-400 text-sm mt-1">Histórico de acessos recentes à sua conta.</p>
            </div>
            
            <div className="overflow-x-auto">
                <table className="w-full text-left text-sm text-slate-400">
                    <thead className="bg-slate-950 text-slate-200 uppercase font-medium">
                        <tr>
                            <th className="px-6 py-4">Dispositivo</th>
                            <th className="px-6 py-4">IP</th>
                            <th className="px-6 py-4">Expira em</th>
                            <th className="px-6 py-4">Status</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800">
                        {sessions.map((session) => (
                            <tr key={session.id} className="hover:bg-slate-800/50 transition-colors">
                                <td className="px-6 py-4 flex items-center gap-3 text-white">
                                    {session.device_name.includes('Postman') ? (
                                        <div className="bg-orange-500/10 p-2 rounded text-orange-500"><Monitor size={18}/></div>
                                    ) : (
                                        <div className="bg-blue-500/10 p-2 rounded text-blue-500"><Smartphone size={18}/></div>
                                    )}
                                    <span className="font-medium truncate max-w-[150px]" title={session.device_name}>
                                        {session.device_name}
                                    </span>
                                </td>
                                <td className="px-6 py-4 font-mono text-slate-500">
                                    {session.ip_address}
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-2">
                                        <Clock size={14} />
                                        {formatDate(session.expires_at)}
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    {session.is_revoked ? (
                                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                                            <XCircle size={12} /> Revogado
                                        </span>
                                    ) : (
                                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-500/10 text-green-400 border border-green-500/20">
                                            <CheckCircle size={12} /> Ativo
                                        </span>
                                    )}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
                {sessions.length === 0 && (
                    <div className="p-8 text-center text-slate-500">Nenhuma sessão encontrada.</div>
                )}
            </div>
        </div>

      </div>
    </div>
  );
}