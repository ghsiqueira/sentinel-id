import { useEffect, useState } from 'react';
import api from '../services/api';
import { useNavigate } from 'react-router-dom';
import { Shield, ArrowLeft, Clock, Activity, AlertTriangle } from 'lucide-react';

interface AuditLog {
  id: string;
  action: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
}

export default function AuditLogs() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    api.get<AuditLog[]>('/users/audit-logs')
      .then(response => {
        setLogs(response.data || []);
      })
      .catch(err => console.error("Erro ao carregar logs", err))
      .finally(() => setLoading(false));
  }, []);

  const getActionColor = (action: string) => {
    if (action.includes('KILL_SWITCH')) return 'text-red-400';
    if (action.includes('LOGIN')) return 'text-green-400';
    return 'text-blue-400';
  };

  const formatAction = (action: string) => {
    return action.replace(/_/g, ' ');
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-6">
      <div className="max-w-4xl mx-auto">
        
        <header className="flex items-center gap-4 mb-8 border-b border-slate-800 pb-6">
          <button 
            onClick={() => navigate('/dashboard')}
            className="p-2 hover:bg-slate-800 rounded-full transition-colors"
          >
            <ArrowLeft size={24} className="text-slate-400" />
          </button>
          <div>
            <h1 className="text-2xl font-bold font-mono tracking-wider flex items-center gap-3">
              <Activity className="text-blue-500" />
              HISTÓRICO DE AUDITORIA
            </h1>
            <p className="text-xs text-slate-400 uppercase tracking-widest mt-1">
              Registro imutável de segurança
            </p>
          </div>
        </header>

        {loading ? (
          <div className="flex justify-center p-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
          </div>
        ) : (
          <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-xl">
            {logs.length === 0 ? (
              <div className="p-12 text-center text-slate-500 flex flex-col items-center gap-4">
                <Shield size={48} className="opacity-20" />
                <p>Nenhum registro de atividade encontrado.</p>
              </div>
            ) : (
              <div className="divide-y divide-slate-800">
                {logs.map((log) => (
                  <div key={log.id} className="p-4 hover:bg-slate-800/30 transition-colors flex flex-col sm:flex-row justify-between gap-4">
                    
                    <div className="flex items-start gap-4">
                      <div className={`p-2 rounded-lg bg-slate-950 border border-slate-800 mt-1`}>
                        {log.action.includes('KILL') ? <AlertTriangle size={20} className="text-red-500"/> : <Shield size={20} className="text-green-500"/>}
                      </div>
                      <div>
                        <p className={`font-bold font-mono text-sm ${getActionColor(log.action)}`}>
                          {formatAction(log.action)}
                        </p>
                        <p className="text-xs text-slate-400 mt-1 font-mono">
                          IP: <span className="text-slate-300">{log.ip_address}</span>
                        </p>
                        <p className="text-[10px] text-slate-500 mt-1 truncate max-w-xs sm:max-w-md" title={log.user_agent}>
                          {log.user_agent}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 text-xs text-slate-500 sm:text-right font-mono bg-slate-950/50 px-3 py-1 rounded h-fit self-start sm:self-center">
                      <Clock size={12} />
                      {new Date(log.created_at).toLocaleString()}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}