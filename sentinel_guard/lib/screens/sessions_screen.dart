import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import 'package:google_fonts/google_fonts.dart';
import '../models/session.dart';
import '../services/api_service.dart';

class SessionsScreen extends StatefulWidget {
  const SessionsScreen({super.key});

  @override
  State<SessionsScreen> createState() => _SessionsScreenState();
}

class _SessionsScreenState extends State<SessionsScreen> {
  final _api = ApiService().dio;
  List<Session> _sessions = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _fetchSessions();
  }

  Future<void> _fetchSessions() async {
    setState(() => _isLoading = true);
    try {
      final response = await _api.get('/users/sessions');
      if (response.data == null) {
        setState(() {
          _sessions = [];
          _isLoading = false;
        });
        return;
      }
      final List<dynamic> data = response.data;
      setState(() {
        _sessions = data.map((json) => Session.fromJson(json)).toList();
        _isLoading = false;
      });
    } on DioException catch (e) {
      setState(() {
        _error = 'Erro de conexão: ${e.message}';
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Erro interno.';
        _isLoading = false;
      });
    }
  }

  Future<void> _revokeSession(Session session) async {
    final isMe = session.isCurrent;

    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF0F172A),
        title: Text(
          isMe ? 'Desconectar ESTE dispositivo?' : 'Desconectar dispositivo?',
          style: const TextStyle(color: Colors.white),
        ),
        content: Text(
          isMe
              ? 'Você será deslogado imediatamente do aplicativo.'
              : 'Este dispositivo perderá o acesso.',
          style: const TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancelar'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text(
              'Desconectar',
              style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    try {
      setState(() {
        _sessions.removeWhere((s) => s.id == session.id);
      });

      await _api.delete('/users/sessions/${session.id}');

      if (isMe) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Desconectado com sucesso.'),
              backgroundColor: Colors.orange,
            ),
          );
          navigatorKey.currentState?.pushNamedAndRemoveUntil(
            '/login',
            (route) => false,
          );
        }
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Sessão desconectada com sucesso!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      _fetchSessions();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Erro ao desconectar.'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Sessões Ativas',
          style: GoogleFonts.inter(fontWeight: FontWeight.bold),
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
          ? Center(
              child: Text(_error!, style: const TextStyle(color: Colors.red)),
            )
          : _sessions.isEmpty
          ? const Center(child: Text("Nenhuma sessão ativa encontrada."))
          : RefreshIndicator(
              onRefresh: _fetchSessions,
              color: Colors.blueAccent,
              backgroundColor: const Color(0xFF0F172A),
              child: ListView.builder(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(16),
                itemCount: _sessions.length,
                itemBuilder: (context, index) {
                  final session = _sessions[index];
                  final deviceInfo = session.deviceInfo.isNotEmpty
                      ? session.deviceInfo
                      : "Dispositivo Desconhecido";
                  final isMobile =
                      deviceInfo.toLowerCase().contains('mobile') ||
                      deviceInfo.toLowerCase().contains('android');

                  return Card(
                    color: session.isCurrent
                        ? const Color(0xFF1E293B)
                        : const Color(0xFF0F172A),
                    margin: const EdgeInsets.only(bottom: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                      side: BorderSide(
                        color: session.isCurrent
                            ? Colors.green.withValues(alpha: 0.5)
                            : Colors.blue.withValues(alpha: 0.1),
                      ),
                    ),
                    child: ListTile(
                      leading: Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: Colors.black26,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Icon(
                          isMobile ? Icons.smartphone : Icons.computer,
                          color: session.isCurrent
                              ? Colors.greenAccent
                              : Colors.blueAccent,
                        ),
                      ),
                      title: Row(
                        children: [
                          Flexible(
                            child: Text(
                              deviceInfo,
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                                fontSize: 14,
                              ),
                            ),
                          ),
                          if (session.isCurrent)
                            Container(
                              margin: const EdgeInsets.only(left: 8),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 6,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.green.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: const Text(
                                "VOCÊ",
                                style: TextStyle(
                                  fontSize: 10,
                                  color: Colors.greenAccent,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                        ],
                      ),
                      subtitle: Text(
                        'IP: ${session.ipAddress}',
                        style: const TextStyle(
                          color: Colors.grey,
                          fontSize: 12,
                        ),
                      ),
                      trailing: IconButton(
                        icon: Icon(
                          Icons.delete_outline,
                          color: session.isCurrent
                              ? Colors.orangeAccent
                              : Colors.redAccent,
                        ),
                        onPressed: () => _revokeSession(session),
                      ),
                    ),
                  );
                },
              ),
            ),
    );
  }
}
