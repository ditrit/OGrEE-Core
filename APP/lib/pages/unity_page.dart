import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_inappwebview/flutter_inappwebview.dart';
import 'package:ogree_app/common/appbar.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/pages/projects_page.dart';
import 'package:url_launcher/url_launcher.dart';

// Demo for Unity WebGL
class UnityPage extends StatefulWidget {
  final String userEmail;
  final Tenant? tenant;
  const UnityPage({super.key, required this.userEmail, this.tenant});

  @override
  State<UnityPage> createState() => UnityPageState();

  static UnityPageState? of(BuildContext context) =>
      context.findAncestorStateOfType<UnityPageState>();
}

class UnityPageState extends State<UnityPage> with TickerProviderStateMixin {
  final GlobalKey webViewKey = GlobalKey();

  InAppWebViewController? webViewController;
  InAppWebViewSettings settings = InAppWebViewSettings(
    isInspectable: kDebugMode,
    mediaPlaybackRequiresUserGesture: false,
    allowsInlineMediaPlayback: true,
    iframeAllow: "camera; microphone",
    iframeAllowFullscreen: true,
  );

  PullToRefreshController? pullToRefreshController;
  String url = "";
  double progress = 0;
  final urlController = TextEditingController();

  @override
  void initState() {
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color.fromARGB(255, 238, 238, 241),
      appBar: myAppBar(
        context,
        widget.userEmail,
        isTenantMode: widget.tenant != null,
      ),
      body: SafeArea(
        child: Column(
          children: <Widget>[
            Expanded(
              child: Stack(
                children: [
                  InAppWebView(
                    key: webViewKey,
                    initialUrlRequest: URLRequest(
                      url: WebUri("https://unity.kube.ogree.ditrit.io"),
                    ),
                    initialSettings: settings,
                    pullToRefreshController: pullToRefreshController,
                    onWebViewCreated: (controller) {
                      webViewController = controller;
                    },
                    onLoadStart: (controller, url) {
                      setState(() {
                        this.url = url.toString();
                        urlController.text = this.url;
                      });
                    },
                    onPermissionRequest: (controller, request) async {
                      return PermissionResponse(
                        resources: request.resources,
                        action: PermissionResponseAction.GRANT,
                      );
                    },
                    shouldOverrideUrlLoading:
                        (controller, navigationAction) async {
                      final uri = navigationAction.request.url!;

                      if (![
                        "http",
                        "https",
                        "file",
                        "chrome",
                        "data",
                        "javascript",
                        "about",
                      ].contains(uri.scheme)) {
                        if (await canLaunchUrl(uri)) {
                          // Launch the App
                          await launchUrl(
                            uri,
                          );
                          // and cancel the request
                          return NavigationActionPolicy.CANCEL;
                        }
                      }

                      return NavigationActionPolicy.ALLOW;
                    },
                    onLoadStop: (controller, url) async {
                      pullToRefreshController?.endRefreshing();
                      setState(() {
                        this.url = url.toString();
                        urlController.text = this.url;
                      });
                    },
                    onReceivedError: (controller, request, error) {
                      pullToRefreshController?.endRefreshing();
                    },
                    onProgressChanged: (controller, progress) {
                      if (progress == 100) {
                        pullToRefreshController?.endRefreshing();
                      }
                      setState(() {
                        this.progress = progress / 100;
                        urlController.text = url;
                      });
                    },
                    onUpdateVisitedHistory: (controller, url, androidIsReload) {
                      setState(() {
                        this.url = url.toString();
                        urlController.text = this.url;
                      });
                    },
                    onConsoleMessage: (controller, consoleMessage) {
                      if (kDebugMode) {
                        print(consoleMessage);
                      }
                    },
                  ),
                  if (progress < 1.0)
                    LinearProgressIndicator(value: progress)
                  else
                    Container(),
                ],
              ),
            ),
            OverflowBar(
              alignment: MainAxisAlignment.center,
              spacing: 10,
              children: <Widget>[
                ElevatedButton(
                  child: const Icon(Icons.arrow_back),
                  onPressed: () => Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (context) => ProjectsPage(
                        userEmail: widget.userEmail,
                        isTenantMode: widget.tenant != null,
                      ),
                    ),
                  ),
                ),
                ElevatedButton(
                  child: const Icon(Icons.refresh),
                  onPressed: () {
                    webViewController?.reload();
                  },
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
