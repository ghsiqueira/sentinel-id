import 'dart:io';
import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:google_fonts/google_fonts.dart';
import '../models/session.dart';
import 'login_screen.dart';

class SessionsScreen extends StatefulWidget {
  const SessionsScreen({super.key});

  @override
  State<SessionsScreen> createState() => _SessionsScreenState();
}

class _SessionsScreenState extends State<SessionsScreen> {
  final _dio = Dio();
  final _storage = const FlutterSecureStorage();

  List<Session> _sessions = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _fetchSessions();
  }

  Future<void> _fetchSessions() async {
    try {
      final token = await _storage.read(key: 'access_token');

      String baseUrl;
      if (Platform.isAndroid) {
        baseUrl = 'http://10.0.2.2:8080';
      } else {
        baseUrl = 'http://localhost:8080';
      }

      final response = await _dio.get(
        '$baseUrl/api/users/sessions',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );

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
      if (e.response?.statusCode == 401) {
        if (mounted) {
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (_) => const LoginScreen()),
          );
        }
      }
      setState(() {
        _error = 'Erro de conexão: ${e.response?.statusCode ?? "Rede"}';
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Erro interno: $e';
        _isLoading = false;
      });
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
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(
                      Icons.error_outline,
                      size: 48,
                      color: Colors.red,
                    ),
                    const SizedBox(height: 16),
                    Text(
                      _error!,
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: Colors.red),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: () {
                        setState(() {
                          _isLoading = true;
                          _error = null;
                        });
                        _fetchSessions();
                      },
                      child: const Text("Tentar Novamente"),
                    ),
                  ],
                ),
              ),
            )
          : _sessions.isEmpty
          ? const Center(child: Text("Nenhuma sessão ativa encontrada."))
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _sessions.length,
              itemBuilder: (context, index) {
                final session = _sessions[index];
                final deviceInfo = session.deviceInfo.isNotEmpty
                    ? session.deviceInfo
                    : "Dispositivo Desconhecido";
                final isMobile =
                    deviceInfo.toLowerCase().contains('android') ||
                    deviceInfo.toLowerCase().contains('ios');

                return Card(
                  color: Colors.blue.withValues(alpha: 0.1),
                  margin: const EdgeInsets.only(bottom: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                    side: BorderSide(color: Colors.blue.withValues(alpha: 0.3)),
                  ),
                  child: ListTile(
                    leading: Icon(
                      isMobile ? Icons.smartphone : Icons.computer,
                      color: Colors.blue,
                      size: 32,
                    ),
                    title: Text(
                      deviceInfo,
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    subtitle: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const SizedBox(height: 4),
                        Text(
                          'IP: ${session.ipAddress}',
                          style: const TextStyle(color: Colors.grey),
                        ),
                        Text(
                          'Criado em: ${session.createdAt.toLocal().toString().split('.')[0]}',
                          style: const TextStyle(
                            color: Colors.grey,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
    );
  }
}
