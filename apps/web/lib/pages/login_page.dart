import 'package:jaspr/dom.dart';
import 'package:jaspr/jaspr.dart';
import 'package:web/web.dart' as web;

Component _t(String value) => Component.text(value);

/// Acceso con correo y contraseña (mismo contrato que la app móvil). En modo demo el formulario viene precargado.
class LoginPage extends StatelessComponent {
  const LoginPage({required this.demo, required this.loading, required this.error, required this.onSubmit, super.key});

  final bool demo;
  final bool loading;
  final String? error;
  final void Function(String email, String password) onSubmit;

  @override
  Component build(BuildContext context) => main_(id: 'contenido', classes: 'login', [
    p(classes: 'kicker', [_t('SOPHIA AI')]),
    h1([_t('Hola, soy Sofía')]),
    p(classes: 'lead', [
      _t(demo ? 'Esta es una demo con datos ficticios. Entra para conversar.' : 'Entra con tu cuenta para continuar tu conversación.'),
    ]),
    if (error != null) p(classes: 'alert', attributes: {'role': 'alert'}, [span([_t(error!)])]),
    form(
      classes: 'login',
      attributes: {'novalidate': 'true'},
      events: {
        'submit': (event) {
          event.preventDefault();
          final email = (web.document.getElementById('email') as web.HTMLInputElement).value.trim();
          final password = (web.document.getElementById('password') as web.HTMLInputElement).value;
          onSubmit(email, password);
        },
      },
      [
        div(classes: 'field', [
          label(attributes: {'for': 'email'}, [_t('Correo')]),
          input(id: 'email', type: InputType.email, name: 'email', value: demo ? 'alex@example.test' : null, attributes: {'autocomplete': 'username', 'required': 'true'}),
        ]),
        div(classes: 'field', [
          label(attributes: {'for': 'password'}, [_t('Contraseña')]),
          input(id: 'password', type: InputType.password, name: 'password', value: demo ? 'demo-password' : null, attributes: {'autocomplete': 'current-password', 'required': 'true'}),
        ]),
        button([_t(loading ? 'Entrando…' : 'Entrar')], type: ButtonType.submit, classes: 'btn btn-primary', disabled: loading),
      ],
    ),
  ]);
}
