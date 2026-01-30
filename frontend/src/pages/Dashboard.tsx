import { useEffect, useState } from 'react';
import api from '../services/api';
import { useNavigate } from 'react-router-dom';
import { Shield, Smartphone, Monitor, LogOut, AlertTriangle, RefreshCw, XCircle } from 'lucide-react';

interface Session {
  id: string;
  device_info: string;
  ip_address: string;
  created_at: string;
  expires_at: string;
  is_revoked: boolean;
}

export default function Dashboard() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [revoking, setRevoking] = useState(false);
  const navigate = useNavigate();

  const fetchSessions = async () => {
    try {
      const response = await api.get<Session[]>('/users/sessions');
      setSessions(Array.isArray(response.data) ? response.data : []);
      setError('');
    } catch (err) {
      console.error(err);
      setError('Falha ao carregar sessões.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSessions();
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
    navigate('/login');
  };

  const handleRevokeAll = async () => {
    if (!confirm('TEM CERTEZA? Isso vai desconectar você e TODOS os outros dispositivos.')) return;

    setRevoking(true);
    try {
      await api.post('/users/revoke-all');
      alert('Todas as sessões foram derrubadas com sucesso!');
      handleLogout();
    } catch (err) {
      console.error(err);
      alert('Erro ao tentar revogar sessões.');
      setRevoking(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center text-white">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-6">
      <div className="max-w-4xl mx-auto">
        
        <header className="flex justify-between items-center mb-10 border-b border-slate-800 pb-6">
          <div className="flex items-center gap-3">
            <Shield className="w-10 h-10 text-blue-500" />
            <div>
              <h1 className="text-2xl font-bold font-mono tracking-wider">SENTINEL ID</h1>
              <p className="text-xs text-slate-400 uppercase tracking-widest">Painel de Controle</p>
            </div>
          </div>
          <button 
            onClick={handleLogout}
            className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors"
          >
            <LogOut size={20} />
            <span>Sair</span>
          </button>
        </header>

        {error && (
          <div className="bg-red-500/10 border border-red-500/50 text-red-400 p-4 rounded-xl mb-6 flex items-center gap-3">
            <XCircle size={24} />
            <span>{error}</span>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="bg-slate-900/50 border border-slate-800 p-6 rounded-xl flex items-center gap-4">
            <div className="p-3 bg-green-500/10 rounded-full">
              <Shield className="w-6 h-6 text-green-500" />
            </div>
            <div>
              <p className="text-sm text-slate-400">Status da Conta</p>
              <p className="text-lg font-bold text-green-400">Protegida & Ativa</p>
            </div>
          </div>

          <div className="bg-slate-900/50 border border-slate-800 p-6 rounded-xl flex items-center gap-4">
            <div className="p-3 bg-blue-500/10 rounded-full">
              <Monitor className="w-6 h-6 text-blue-500" />
            </div>
            <div>
              <p className="text-sm text-slate-400">Sessões Ativas</p>
              <p className="text-lg font-bold text-white">{sessions.length} dispositivos</p>
            </div>
          </div>
        </div>

        <div className="bg-red-500/5 border border-red-500/20 rounded-xl p-8 mb-10 text-center">
          <div className="inline-flex p-4 bg-red-500/10 rounded-full mb-4">
            <AlertTriangle className="w-10 h-10 text-red-500" />
          </div>
          <h2 className="text-xl font-bold text-red-400 mb-2">ZONA DE PERIGO</h2>
          <p className="text-slate-400 mb-6 max-w-md mx-auto">
            Se você suspeita que sua conta foi comprometida, use o botão abaixo para invalidar 
            todos os tokens de acesso imediatamente.
          </p>
          <button
            onClick={handleRevokeAll}
            disabled={revoking}
            className="bg-red-600 hover:bg-red-700 text-white px-8 py-3 rounded-lg font-bold transition-all shadow-lg shadow-red-900/20 disabled:opacity-50"
          >
            {revoking ? 'DERRUBANDO...' : 'REVOGAR ACESSO GLOBAL (KILL SWITCH)'}
          </button>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
          <div className="p-6 border-b border-slate-800 flex justify-between items-center">
            <h3 className="font-bold text-lg">Dispositivos Conectados</h3>
            <button onClick={fetchSessions} className="p-2 hover:bg-slate-800 rounded-lg transition-colors">
              <RefreshCw size={18} className="text-slate-400" />
            </button>
          </div>

          <div className="divide-y divide-slate-800">
            {sessions.length === 0 ? (
              <div className="p-8 text-center text-slate-500">Nenhuma sessão encontrada.</div>
            ) : (
              sessions.map((session) => {
                const deviceInfo = session.device_info || '';
                const isMobile = /android|iphone|ipad|mobile/i.test(deviceInfo);
                const date = new Date(session.created_at).toLocaleString();

                return (
                  <div key={session.id} className="p-4 flex items-center justify-between hover:bg-slate-800/50 transition-colors">
                    <div className="flex items-center gap-4">
                      <div className={`p-3 rounded-lg ${isMobile ? 'bg-purple-500/10 text-purple-400' : 'bg-blue-500/10 text-blue-400'}`}>
                        {isMobile ? <Smartphone size={24} /> : <Monitor size={24} />}
                      </div>
                      <div>
                        <p className="font-medium text-white">
                          {deviceInfo || 'Dispositivo Desconhecido'}
                        </p>
                        <div className="flex gap-3 text-xs text-slate-400 mt-1">
                          <span>IP: {session.ip_address}</span>
                          <span>•</span>
                          <span>{date}</span>
                        </div>
                      </div>
                    </div>
                    <div className="text-xs font-mono text-slate-500 bg-slate-950 px-2 py-1 rounded border border-slate-800">
                      ATIVO
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

      </div>
    </div>
  );
}