English version below ...

### Zweck

Dieses Programm ermöglicht die Interaktion mit Google Gemini AI, um Prompts zu senden und Antworten in verschiedenen Formaten (Markdown, ANSI, HTML) zu empfangen. Das Ziel ist die nahtlose Integration von KI-Funktionen in den Arbeitsablauf des Benutzers durch die Unterstützung verschiedener Eingabe- und Ausgabekanäle sowie einer Historienverwaltung.

Das Programm ist für BSD, Linux, macOS und Windows verfügbar.

### Google Gemini KI

Dieses Programm erlaubt die Nutzung verschiedener Modelle der Google Gemini KI-Familie. Viele Parameter des KI-Modells, wie z.B. die Anzahl der Antworten oder die Varianz, können in der Konfiguration angepasst werden.

### Installation, Konfiguration

Die Anwendung enthält intern alle für die Nutzung notwendigen Komponenten. Fehlende Komponenten werden beim Start automatisch installiert. Daher ist es ausreichend, die Anwendung in ein beliebiges Verzeichnis zu kopieren und zu starten. Dies ermöglicht eine einfache projektbezogene Nutzung der Anwendung. Die Anwendung kann umfangreich über eine YAML-Datei konfiguriert werden.

### Eingabe der Abfragen

Abfragen können über verschiedene Kanäle eingegeben werden: direkt im Terminal, über die Textdatei 'prompt-input.txt', oder über 'localhost' (Port 4242). Für eine komfortablere Prompterstellung und -ausführung kann die Webseite 'prompt-input.html' verwendet werden.

### Ausgabe der Abfrage+Antwort-Paare

Die Abfrage+Antwort-Paare werden in verschiedenen Formaten ausgegeben: im Terminal (ANSI-farbig), als Markdown-Dateien (für Editoren/Viewer) und als HTML-Seiten (für Browser). Die Browser-Ausgabe bietet umfangreiche Anpassungs- und Nutzungsmöglichkeiten.

### Benachrichtigungen

Benachrichtigungen informieren optional über den Start und Abschluss der Prompt-Verarbeitung, weil die Antwortgenerierung durch KI einige Zeit in Anspruch nehmen kann.

### Nutzung und Hinweise

Dieses Programm kann im Terminal für direkte Eingabe und Ausgabe genutzt werden oder als Controller mit Eingabe per Datei/localhost und Ausgabe in GUI-Programmen.

**Wichtig**: Diese Anwendung dient der Evaluierung und ist zum Testen von Google Gemini AI mit einem persönlichen Gemini API-Key gedacht. Beachten Sie die Hinweise zu Nutzungsbedingungen und Datenschutz im folgenden Abschnitt. 

### Nutzungsbedingungen und Datenschutzhinweise

Beachten Sie die [Google-Nutzungsbedingungen](https://policies.google.com/terms) und die [Richtlinie zur unzulässigen Nutzung von generativer KI](https://policies.google.com/terms/generative-ai/use-policy) für Gemini. Weitere Informationen zum Datenschutz finden Sie im [Gemini-Apps Privacy Hub](https://support.google.com/gemini?p=privacy_help).

**Wichtig**: Die frei verfügbare Version von Google Gemini AI verwendet Ihre Eingabe- und Ausgabedaten zur Verbesserung des Modells. Verarbeiten Sie daher keine privaten oder vertraulichen Daten. Für die kostenpflichtige Version gelten andere Datenschutzrichtlinien.

Siehe auch [häufig gestellte Fragen zu Gemini-Apps](https://gemini.google.com/faq?hl=de).

### Technische Hinweise

Jedes Gemini-KI-Modell hat ein maximales Output-Token-Limit, welches die Antwortlänge begrenzt. Ein Token entspricht dabei grob zirka 4 Zeichen, 100 Token in etwa 60 bis 80 englischen Wörtern. Wenn das KI-Modell das Output-Token-Limit beispielsweise auf 8192 Tokens definiert, entspricht dies einer maximalen Antwortlänge von 4900 bis 6500 englischen Wörtern.

### Voraussetzungen und Konfiguration

Für die Nutzung ist ein persönlicher Gemini API-Key von Google erforderlich. Konfigurieren Sie den API-Key als Umgebungsvariable ```GEMINI_API_KEY```. Für die Internetverbindung kann optional ein Proxy konfiguriert werden.

* Konfiguration des API-Keys im Programm-Environment:
  * macOS, Linux: ```export GEMINI_API_KEY=Your-API-Key```
  * Windows: ```setx GEMINI_API_KEY Your-API-Key``` (erfordert Neustart des Terminals)

### Support und Programme

Programmfehler bitte in 'Issues' melden, Diskussionen und Fragen in 'Discussions'. Ausführbare Programme finden Sie im 'Releases'-Bereich.

***

![overview diagram](./gem-pro.png)

***

### Purpose

This program enables interaction with a Google Gemini AI model to send prompts and receive responses in various formats (Markdown, ANSI, HTML). It aims to seamlessly integrate AI capabilities into the user's workflow by supporting various input and output channels along with a history feature..

The program is available for BSD, Linux, macOS, and Windows.

### Google Gemini AI

This program allows the use of various models from the Google Gemini AI family. Many parameters of the AI model, such as the number of responses or variance, can be adjusted in the configuration.

### Installation, Configuration

The application internally includes all components required for use. Missing components are automatically installed when the application is started. Therefore, copying the application to a directory and starting it is sufficient. This allows for simple project-related use. The application can be extensively configured via a YAML file.

### Input of Prompts

Prompts can be entered via various channels: directly in the terminal, via the text file 'prompt-input.txt', or through 'localhost' (Port 4242). For more convenient prompt creation and execution, the webpage 'prompt-input.html' can be used.

### Output of Prompt+Response Pairs

The output of prompt+response pairs is available in various formats: in the terminal (ANSI colored), as Markdown files (for editors/viewers), and as HTML pages (for browsers). The browser output offers extensive customization and usage options.

### Notifications

Optional notifications inform users about the start and completion of prompt processing, as AI response generation can take some time.

### Usage and Notes

This program can be used in a terminal for direct input and output or as a controller with input via file/localhost and output in GUI programs.

Important: This application is for evaluation purposes and intended as a trial of Google Gemini AI with a personal Gemini API key. Please note the terms of service and privacy information in the following section.

### Terms of Service and Privacy Notes

Refer to the [Google Terms of Service](https://policies.google.com/terms) and the [Generative AI Prohibited Use Policy](https://policies.google.com/terms/generative-ai/use-policy) for Gemini. For more information on privacy, please visit the [Gemini Apps Privacy Hub](https://support.google.com/gemini?p=privacy_help).

**Important**: The freely available version of Google Gemini AI uses your input and output data to improve the model. Therefore, do not process any private or confidential data. Different privacy policies apply to the paid version.

See also [Gemini Apps FAQ](https://gemini.google.com/faq?hl=en).

### Technical Notes

Each Gemini AI model has a maximum output token limit, which restricts the response length. One token roughly corresponds to approximately 4 characters, 100 tokens to about 60 to 80 English words. If, for example, the AI model defines the output token limit as 8192 tokens, that corresponds to a maximum response length of 4900 to 6500 English words.

### Required Prerequisites

A personal Gemini API key from Google is required for use. Configure the API key as an environment variable ```GEMINI_API_KEY```. An internet proxy can be optionally configured for the internet connection.

* Configure the API key in the program environment:
  * macOS, Linux: ```export GEMINI_API_KEY=Your-API-Key```
  * Windows: ```setx GEMINI_API_KEY Your-API-Key``` (requires terminal restart)

### Support and Programs

Please report program errors in 'Issues'. For discussions and questions, please use 'Discussions'. Executable programs are available in the 'Releases' section.
