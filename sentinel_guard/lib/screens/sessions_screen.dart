import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../services/api_service.dart';

class SessionsScreen extends StatefulWidget {
  const SessionsScreen({super.key});

  @override
  State<SessionsScreen> createState() => _SessionsScreenState();
}

class _SessionsScreenState extends State<SessionsScreen> {
  final _api = ApiService().dio;
  bool _isLoading = true;
  List<dynamic> _sessions = [];

  @override
  void initState() {
    super.initState();
    _fetchSessions();
  }

  Future<void> _fetchSessions() async {
    setState(() => _isLoading = true);
    try {
      final response = await _api.get('/users/sessions');

      if (mounted) {
        setState(() {
          if (response.data is List) {
            _sessions = response.data;
          } else if (response.data is Map &&
              response.data.containsKey('sessions')) {
            _sessions = response.data['sessions'];
          } else {
            _sessions = [];
          }
          _isLoading = false;
        });
      }
    } catch (e) {
      debugPrint('Erro real ao buscar sessões: $e');
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Erro ao carregar sessões.'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _revokeSession(String sessionId) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF1E293B),
        title: const Text(
          'Revogar Sessão',
          style: TextStyle(color: Colors.white),
        ),
        content: const Text(
          'Tem a certeza que deseja desconectar este dispositivo?',
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
              'Sim, Desconectar',
              style: TextStyle(color: Colors.red),
            ),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    try {
      await _api.delete('/users/sessions/$sessionId');

      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Sessão encerrada com sucesso!'),
          backgroundColor: Colors.green,
        ),
      );
      _fetchSessions();
    } catch (e) {
      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Erro ao encerrar sessão.'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  String _getDeviceFriendlyName(String rawDevice) {
    if (rawDevice.isEmpty) return "Dispositivo Desconhecido";
    if (rawDevice.contains('Windows')) return "💻 Windows (Web)";
    if (rawDevice.contains('Macintosh')) return "💻 MacBook / iMac (Web)";
    if (rawDevice.contains('Linux')) return "💻 Linux (Web)";
    if (rawDevice.contains('Android')) return "📱 Android";
    if (rawDevice.contains('iPhone') || rawDevice.contains('iPad'))
      return "📱 iOS";
    if (rawDevice.contains('Dart')) return "📱 Sentinel Guard App";
    return "💻 Dispositivo Web";
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        title: Text(
          'Dispositivos Conectados',
          style: GoogleFonts.jetBrainsMono(
            fontWeight: FontWeight.bold,
            fontSize: 18,
          ),
        ),
        centerTitle: true,
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Colors.blue))
          : _sessions.isEmpty
          ? const Center(
              child: Text(
                'Nenhuma sessão ativa.',
                style: TextStyle(color: Colors.grey),
              ),
            )
          : RefreshIndicator(
              onRefresh: _fetchSessions,
              color: Colors.blue,
              backgroundColor: const Color(0xFF1E293B),
              child: ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: _sessions.length,
                itemBuilder: (context, index) {
                  final session = _sessions[index];
                  final isCurrent = session['is_current'] == true;

                  final rawDevice =
                      session['user_agent'] ?? session['device_info'] ?? '';
                  final rawIp = session['ip_address'] ?? '';

                  final deviceName = _getDeviceFriendlyName(rawDevice);
                  // Puxamos direto do banco cru
                  final location = rawIp.isNotEmpty
                      ? rawIp
                      : "Localização Desconhecida";

                  return Container(
                    margin: const EdgeInsets.only(bottom: 16),
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: const Color(0xFF1E293B),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: isCurrent
                            ? Colors.green.withValues(alpha: 0.5)
                            : Colors.blue.withValues(alpha: 0.2),
                        width: isCurrent ? 2 : 1,
                      ),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              child: Text(
                                deviceName,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontWeight: FontWeight.bold,
                                  fontSize: 16,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                            if (isCurrent)
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.green.withValues(alpha: 0.2),
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: const Text(
                                  'Este Dispositivo',
                                  style: TextStyle(
                                    color: Colors.greenAccent,
                                    fontSize: 12,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              )
                            else
                              IconButton(
                                icon: const Icon(
                                  Icons.delete_outline,
                                  color: Colors.redAccent,
                                ),
                                onPressed: () => _revokeSession(session['id']),
                                constraints: const BoxConstraints(),
                                padding: EdgeInsets.zero,
                              ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        Row(
                          children: [
                            const Icon(
                              Icons.location_on_outlined,
                              color: Colors.grey,
                              size: 16,
                            ),
                            const SizedBox(width: 6),
                            Expanded(
                              child: Text(
                                location,
                                style: const TextStyle(
                                  color: Colors.grey,
                                  fontSize: 13,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 6),
                        Row(
                          children: [
                            const Icon(
                              Icons.access_time,
                              color: Colors.grey,
                              size: 16,
                            ),
                            const SizedBox(width: 6),
                            Text(
                              'Criada em: ${DateTime.parse(session['created_at']).toLocal().toString().split('.')[0]}',
                              style: const TextStyle(
                                color: Colors.grey,
                                fontSize: 13,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
    );
  }
}
