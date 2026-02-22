import { useEffect, useState, useRef } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import QRCode from 'react-qr-code';
import api from '../services/api';
import { ArrowLeft, Smartphone, Loader2, CheckCircle, RefreshCw } from 'lucide-react';

export default function LoginQR() {
  const navigate = useNavigate();
  const [qrData, setQrData] = useState<string | null>(null);
  const [requestId, setRequestId] = useState<string | null>(null);
  const [status, setStatus] = useState<'loading' | 'waiting' | 'approved' | 'expired'>('loading');
  
  const pollInterval = useRef<any>(null);

  const initQR = async () => {
    setStatus('loading');
    setQrData(null);
    try {
      const response = await api.post('/auth/qr/init');
      setRequestId(response.data.request_id);
      
      setQrData(response.data.request_id);
      setStatus('waiting');
    } catch (error) {
      console.error("Erro ao gerar QR", error);
      alert("Erro ao conectar com servidor");
    }
  };

  useEffect(() => {
    initQR();
    return () => stopPolling();
  }, []);

  useEffect(() => {
    if (requestId && status === 'waiting') {
      startPolling();
    }
  }, [requestId, status]);

  const stopPolling = () => {
    if (pollInterval.current) {
      clearInterval(pollInterval.current);
      pollInterval.current = null;
    }
  };

  const startPolling = () => {
    stopPolling(); 

    pollInterval.current = setInterval(async () => {
      if (!requestId) return;

      try {
        const response = await api.get(`/auth/qr/poll/${requestId}`);

        if (response.status === 202) {
           return; 
        }

        if (response.status === 200 && response.data.access_token) {
          stopPolling();
          setStatus('approved');
          
          localStorage.setItem('token', response.data.access_token);
          localStorage.setItem('refresh_token', response.data.refresh_token);
          
          setTimeout(() => navigate('/dashboard'), 1500); 
        }

      } catch (error: any) {
        if (error.response && error.response.status !== 202) {
            stopPolling();
            setStatus('expired');
        }
      }
    }, 2000); 
  };

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-800 p-8 rounded-2xl shadow-2xl w-full max-w-md text-center">
        
        <div className="flex justify-start mb-4">
            <Link to="/login" className="text-slate-400 hover:text-white transition-colors">
                <ArrowLeft />
            </Link>
        </div>

        <h1 className="text-2xl font-bold text-white mb-2">Login sem Senha</h1>
        <p className="text-slate-400 text-sm mb-8">
            Abra o Sentinel App no seu celular e escaneie o código abaixo.
        </p>

        <div className="bg-white p-4 rounded-xl w-fit mx-auto mb-6 relative">
            {status === 'loading' && (
                <div className="w-48 h-48 flex items-center justify-center">
                    <Loader2 className="animate-spin text-slate-900 w-8 h-8" />
                </div>
            )}

            {status === 'waiting' && qrData && (
                <div style={{ height: "auto", margin: "0 auto", maxWidth: 192, width: "100%" }}>
                    <QRCode
                        size={256}
                        style={{ height: "auto", maxWidth: "100%", width: "100%" }}
                        value={qrData}
                        viewBox={`0 0 256 256`}
                    />
                </div>
            )}

            {status === 'approved' && (
                <div className="w-48 h-48 flex flex-col items-center justify-center text-green-600 gap-2">
                    <CheckCircle className="w-16 h-16" />
                    <span className="font-bold">Aprovado!</span>
                </div>
            )}

            {status === 'expired' && (
                <div className="w-48 h-48 flex flex-col items-center justify-center text-red-500 gap-2">
                    <RefreshCw className="w-12 h-12" />
                    <span className="text-sm font-bold">Expirou</span>
                    <button 
                        onClick={initQR}
                        className="text-xs bg-slate-900 text-white px-3 py-1 rounded mt-2 hover:bg-slate-700"
                    >
                        Gerar Novo
                    </button>
                </div>
            )}
        </div>

        <div className="flex items-center justify-center gap-2 text-slate-500 text-sm">
            <Smartphone size={16} />
            <span>Mantenha o app atualizado</span>
        </div>

      </div>
    </div>
  );
}