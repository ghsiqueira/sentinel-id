import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:local_auth/local_auth.dart';
import '../services/api_service.dart';
import 'qr_scanner_screen.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  final _api = ApiService().dio;
  final _storage = const FlutterSecureStorage();
  final LocalAuthentication _localAuth = LocalAuthentication();

  bool _isLoading = false;
  Timer? _statusTimer;

  @override
  void initState() {
    super.initState();
    _startStatusCheck();
  }

  @override
  void dispose() {
    _statusTimer?.cancel();
    super.dispose();
  }

  void _startStatusCheck() {
    _statusTimer = Timer.periodic(const Duration(seconds: 5), (timer) async {
      try {
        await _api.get('/users/me');
      } catch (e) {
        debugPrint('Background check ignorado: $e');
      }
    });
  }

  Future<void> _logout() async {
    _statusTimer?.cancel();
    await _storage.deleteAll();
    if (mounted) {
      Navigator.pushNamedAndRemoveUntil(context, '/login', (route) => false);
    }
  }

  Future<void> _revokeAllSessions() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF0F172A),
        title: const Text(
          'Zona de Perigo',
          style: TextStyle(color: Colors.red),
        ),
        content: const Text(
          'Isso vai desconectar você e TODOS os outros dispositivos imediatamente. Tem certeza?',
          style: TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancelar'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text(
              'SIM, DERRUBAR TUDO',
              style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    setState(() => _isLoading = true);

    try {
      await _api.post('/users/revoke-all');

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('SUCESSO: Todas as sessões foram destruídas!'),
            backgroundColor: Colors.red,
          ),
        );
        _logout();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Erro ao executar comando.'),
            backgroundColor: Colors.orange,
          ),
        );
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _scanAndApproveLogin() async {
    final String? scannedRequestId = await Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => const QRScannerScreen()),
    );

    if (scannedRequestId == null || !mounted) return;

    try {
      final bool canAuthenticateWithBiometrics =
          await _localAuth.canCheckBiometrics;
      final bool canAuthenticate =
          canAuthenticateWithBiometrics || await _localAuth.isDeviceSupported();

      if (canAuthenticate) {
        final bool didAuthenticate = await _localAuth.authenticate(
          localizedReason: 'Confirme a sua identidade para aprovar o login Web',
        );

        if (!didAuthenticate) {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Autenticação cancelada.'),
                backgroundColor: Colors.orange,
              ),
            );
          }
          return;
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Biometria não disponível neste dispositivo.'),
              backgroundColor: Colors.orange,
            ),
          );
        }
        return;
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Erro no leitor biométrico.'),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    setState(() => _isLoading = true);
    try {
      await ApiService().approveQrLogin(scannedRequestId);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Login autorizado com sucesso!'),
            backgroundColor: Colors.green,
            duration: Duration(seconds: 4),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Código QR expirado ou inválido.'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(
          'COMANDO CENTRAL',
          style: GoogleFonts.jetBrainsMono(fontWeight: FontWeight.bold),
        ),
        centerTitle: true,
        actions: [
          IconButton(
            icon: const Icon(Icons.logout, color: Colors.grey),
            onPressed: _logout,
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: Colors.blue.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: Colors.blue.withValues(alpha: 0.3)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.wifi, color: Colors.green),
                  const SizedBox(width: 16),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'STATUS DA REDE',
                        style: TextStyle(color: Colors.grey, fontSize: 12),
                      ),
                      Text(
                        'Conexão Segura',
                        style: GoogleFonts.inter(
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const Spacer(),
            GestureDetector(
              onTap: _isLoading ? null : _revokeAllSessions,
              child: Container(
                width: 200,
                height: 200,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: LinearGradient(
                    colors: [Colors.red.shade900, Colors.red.shade600],
                    begin: Alignment.bottomLeft,
                    end: Alignment.topRight,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.red.withValues(alpha: 0.5),
                      blurRadius: 30,
                      spreadRadius: 5,
                    ),
                  ],
                ),
                child: Center(
                  child: _isLoading
                      ? const CircularProgressIndicator(color: Colors.white)
                      : const Icon(
                          Icons.power_settings_new,
                          size: 80,
                          color: Colors.white,
                        ),
                ),
              ),
            ),
            const SizedBox(height: 32),
            Center(
              child: Text(
                "KILL SWITCH",
                style: GoogleFonts.jetBrainsMono(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            const Center(
              child: Text(
                "Pressione para revogar acesso global",
                style: TextStyle(color: Colors.grey),
              ),
            ),
            const Spacer(),

            OutlinedButton.icon(
              onPressed: () {
                Navigator.pushNamed(context, '/audit-logs');
              },
              icon: const Icon(Icons.history, color: Colors.purpleAccent),
              label: const Text('Ver Histórico de Ações'),
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.purpleAccent,
                side: const BorderSide(color: Colors.purpleAccent),
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
            ),

            const SizedBox(height: 12),

            OutlinedButton.icon(
              onPressed: () {
                Navigator.pushNamed(context, '/sessions');
              },
              icon: const Icon(Icons.devices, color: Colors.blue),
              label: const Text('Ver Dispositivos Conectados'),
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.blue,
                side: const BorderSide(color: Colors.blue),
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
            ),

            const SizedBox(height: 24),

            // NOVO BOTÃO DE APROVAR LOGIN - Fixo na tela, não flutua mais!
            ElevatedButton.icon(
              onPressed: _isLoading ? null : _scanAndApproveLogin,
              icon: const Icon(Icons.qr_code_scanner, color: Colors.white),
              label: const Text(
                'APROVAR LOGIN WEB',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                  letterSpacing: 1.5,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blue.shade600,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }
}
