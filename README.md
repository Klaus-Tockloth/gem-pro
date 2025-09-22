English version below ...

### Zweck

Dieses Programm ermöglicht die Interaktion mit Google Gemini AI, um Prompts zu senden und Antworten in verschiedenen Formaten (Markdown, ANSI, HTML) zu empfangen. Das Ziel ist die nahtlose Integration von KI-Funktionen in den Arbeitsablauf des Benutzers durch die Unterstützung verschiedener Eingabe- und Ausgabekanäle sowie einer Historienverwaltung.

Das Programm ist für BSD, Linux, macOS und Windows verfügbar.

### Google Gemini KI

Dieses Programm erlaubt die Nutzung verschiedener Modelle der Google Gemini KI-Familie. Viele Parameter des KI-Modells, wie z.B. die Anzahl der Antworten oder die Varianz, können in der Konfiguration angepasst werden.

### Google Gemini KI-Modelle

Die Gemini KI-Familie besteht aus folgenden Modellen:

|  | Lite | Flash | Pro |
|---|---|---|---|
| Am Besten geeignet für | viele kosteneffiziente Anfragen | Standardanfragen bei guter Leistung | Programmierung und komplexe Anfragen |
| Geschwindigkeit | hoch | mittel | niedrig |
| Leistung | niedrig | mittel | hoch |
| Kosten | niedrig | mittel | hoch |

### Funktionsumfang

* Mehrere Eingabequellen (Terminal, Datei-Polling, HTTP).
* Unterstützung für Datei-Uploads als Teil des Prompts.
* Konfigurierbare Ausgabeformate (Markdown, ANSI, HTML).
* Speicherung des Verlaufs (History) für jedes Ausgabeformat.
* Chat-Modus für konversationelle Interaktionen.
* Proxy-Unterstützung für Umgebungen ohne direkten Internetzugang.
* Detaillierte Konfigurationsmöglichkeiten über YAML und Kommandozeilen-Flags.
* Anzeige verfügbarer Gemini-Modelle.
* OS-spezifische Konfigurationen (z.B. für externe Viewer-Applikationen).
* Einbettung von Standard-Konfigurationsdateien und Web-Assets für eine einfache Bereitstellung.
* Gute Benutzerführung durch informative Ausgaben (Konfiguration, Modell-Infos, Usage-Text).
* Verarbeitung von Metadaten aus der Gemini-Antwort (Token-Nutzung, Zitate, Grounding-Infos wie Web-Quellen).

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

Hinweise zum Nicht-Chat-Modus (Standard):

* Jeder Prompt wird unabhängig behandelt.
* Die KI erinnert sich nicht an frühere Interaktionen.
* Dateien werden mit jedem Prompt gesendet.

Hinweise zum Chat-Modus (-chatmode Flag):

* Die KI merkt sich den Gesprächsverlauf innerhalb einer Sitzung.
* Dateien werden nur mit dem ersten Prompt gesendet.

**Wichtig**: Diese Anwendung ist zur Nutzung der Google Gemini KI mit einem persönlichen Gemini API-Key gedacht. Beachten Sie die Hinweise zu Nutzungsbedingungen und Datenschutz im folgenden Abschnitt. 

### Nutzungsbedingungen und Datenschutzhinweise

Beachten Sie die [Google-Nutzungsbedingungen](https://policies.google.com/terms) und die [Richtlinie zur unzulässigen Nutzung von generativer KI](https://policies.google.com/terms/generative-ai/use-policy) für Gemini. Weitere Informationen zum Datenschutz finden Sie im [Gemini-Apps Privacy Hub](https://support.google.com/gemini?p=privacy_help).

**Wichtig**: Die frei verfügbare Version von Google Gemini AI verwendet Ihre Eingabe- und Ausgabedaten zur Verbesserung des Modells. Verarbeiten Sie daher keine privaten oder vertraulichen Daten. Für die kostenpflichtige Version gelten andere Datenschutzrichtlinien.

Siehe auch [häufig gestellte Fragen zu Gemini-Apps](https://gemini.google.com/faq?hl=de).

### Technische Hinweise

Jedes Gemini-KI-Modell hat ein maximales Output-Token-Limit, welches die Antwortlänge begrenzt. Ein Token entspricht dabei grob zirka 4 Zeichen, 100 Token in etwa 60 bis 80 englischen Wörtern. Wenn das KI-Modell das Output-Token-Limit beispielsweise auf 65536 Tokens definiert, entspricht dies einer maximalen Antwortlänge von 39300 bis 52400 englischen Wörtern.

### Voraussetzungen und Konfiguration

Für die Nutzung ist ein persönlicher Gemini API-Key von Google erforderlich. Konfigurieren Sie den API-Key als Umgebungsvariable ```GEMINI_API_KEY```. Für die Internetverbindung kann optional ein Proxy konfiguriert werden.

* Konfiguration des API-Keys im Programm-Environment:
  * macOS, Linux: ```export GEMINI_API_KEY=Your-API-Key```
  * Windows: ```setx GEMINI_API_KEY Your-API-Key``` (erfordert Neustart des Terminals)

### Umgang mit Dateien

Abhängig vom Einsatzfall stehen verschiedene Möglichkeiten zur Nutzung von Dateiinhalten im Kontext "Prompt" zur Verfügung.

* Ein textueller Dateiinhalt wird direkt in den Prompt eingefügt.
* Ein Satz an Dateien (1-n) wird über die Kommandozeile definiert:
  * Die Dateien werden anschließend automatisch Bestandteil des Prompts.
* Ein Satz an Dateien (1-n) wird zunächst in den Google-File-Store hochgeladen:
  * Über die Option '-include-files' werden die Dateien anschließend Bestandteil des Prompts.

Die vorgenannten Varianten sind kombinierbar und unterscheiden sich in ihrer Wirkung auf den Prompt nicht. Liegt der zu berücksichtigende Satz an Dateien bereits im Google-File-Store vor, so genügt es im Prompt darauf zu referenzieren. Ein wiederholtes, zeitaufwändiges Hochladen mit jedem neuen Prompt ist somit nicht notwendig. Die Lebensdauer der Dateien im Google-File-Store ist begrenzt. Die Dateien werden nach einer bestimmten Zeit (z.B. 48 Stunden) automatisch gelöscht. Der Satz an Dateien im Google-File-Store wird als 'Einheit' betrachtet und immer vollständig im Prompt referenziert.

### Umgang mit einem expliziten Cache

Ein Cache beinhaltet 1-n Dateien. Die Dateien werden bei der Erzeugung des Caches tokenisiert. Ein Cache bietet somit zwei Vorteile:

* Im Prompt kann auf den Cache referenziert werden.
* Die aufwändige Tokenisierung erfolgt nur einmal.

Je größer der Cache ist, desto vorteilhafter ist dessen Nutzung. Ein Cache ist kostenpflicht. Richtig eingesetzt sind schnellere Antwortzeiten und geringere Gesamtkosten möglich. Die Lebensdauer eines Caches kann konfiguriert werden (z.B. 4 h). Das gewählte KI-Modell muss 'Caching' unterstützen. Beim Anlegen wird der Cache vom aktuell gewählten KI-Modell tokenisiert. Anschließend kann der Cache auch nur von diesem KI-Modell genutzt werden.

In einem Prompt kann nur ein Cache referenziert werden kann. Alle Cache-Operationen wirken deshalb immer auf einen modell-spezifischen Cache mit dem logischen Namen "gem-pro-cache" (konfigurierbar). Dieser wird über die Option '-include-cache' in den Prompt übernommen. 

### Impliziter Cache im Chat-Modus

Eine hervorragende Möglichkeit des (impliziten) Cachings bietet der Chat-Modus. Hier werden die Dateien nur bei Eröffnung einer Session, also als Bestandteil des ersten Prompts, übertragen und tokenisiert. Gemini cached diese Daten, sodass sie bei nachfolgenden Prompts (implizit) zur Verfügung stehen.

### Tooling

Um die Fähigkeiten des KI-Modells zu erweitern, können spezialisierte Werkzeuge (Tools) genutzt werden. Diese ermöglichen es dem Modell, auf externe Informationen zuzugreifen oder komplexe Operationen auszuführen, die über die reine Textgenerierung hinausgehen.

#### Google Search

Ist dieses Tool aktiviert, kann das KI-Modell auf Google Search zugreifen, um aktuelle Informationen zu finden und Fakten zu überprüfen. Dies verbessert die Genauigkeit und Relevanz der Antworten, insbesondere bei Fragen zu aktuellen Ereignissen oder spezifischem Wissen.

#### Code Execution

Dieses Tool erlaubt es dem KI-Modell, Code (in der Regel Python) zu generieren, auszuführen und zu validieren, um eine Lösung für eine Anfrage zu finden. Dies ist besonders nützlich für mathematische Berechnungen, Datenanalysen oder logische Probleme, bei denen eine exakte Ausführung von Schritten erforderlich ist. Das Modell kann so Aufgaben lösen, die präzise und wiederholbare Ergebnisse erfordern.

### Gemini KI-Modelle

Im Kontext dieser Applikation sind folgende KI-Modell-Linien von besonderem Interesse:

* **Flash**-Linie mit den Stärken hohe Geschwindigkeit und Kosteneffizienz, Vielseitigkeit
* **Pro**-Linie mit den Stärken maximale Leistung und Genauigkeit, fortgeschrittenes Schlussfolgern, Programmierunterstützung

Je nach Aufgabenstellung kann auch die Nutzung beider Modell-Linien hilfreich sein:

* Flash-Linie: Generierung mehrerer Lösungsvarianten
* Benutzer: Auswahl der besten Lösungsvariante
* Pro-Linie: Verfeinerung der ausgewählten Lösungsvariante

### Wichtige Parameter

*   **Temperature**: Steuert die Kreativität. Ein niedriger Wert (nahe 0.0) macht die Ausgabe deterministischer und auf die wahrscheinlichsten Wörter fokussiert. Ein hoher Wert (über 0.7) erhöht die Zufälligkeit und fördert vielfältigere, unerwartetere Ergebnisse.

*   **TopP (Nucleus Sampling)**: Filtert die Wortauswahl basierend auf kumulativer Wahrscheinlichkeit. Ein niedriger Wert (z. B. 0.1) beschränkt die Auswahl auf eine sehr kleine Gruppe der wahrscheinlichsten Wörter, was die Antwort fokussiert. Ein hoher Wert (nahe 1.0) erlaubt eine größere Auswahl an Wörtern und erhöht so die Vielfalt.

#### Defaultwerte

Die Standardwerte für die Gemini-Modelle sind in der Regel:
*   **Temperature**: 1.0
*   **TopP**: 0.95

#### Empfohlene Werte für verschiedene Aufgaben

Die nachfolgenden Werte sind Ausgangspunkte. Die ideale Einstellung kann je nach KI-Modell und der genauen Aufgabenstellung variieren.

| Aufgabe | Temperature | TopP | Beschreibung |
| :--- | :--- | :--- | :--- |
| **Allgemeine Aufgaben** | ~0.7 | ~0.95 | Ein guter Mittelweg für ausgewogene und kohärente, aber nicht zu starre Antworten. |
| **Kreative Aufgaben** | 0.7 - 1.0+ | 0.9 - 0.99 | Fördert neuartige Ideen, Brainstorming und die Erstellung einzigartiger Texte wie Gedichte oder Geschichten. |
| **Fakten & Logik** | 0.0 - 0.4 | < 0.8 | Maximiert Genauigkeit und Faktentreue für Erklärungen oder mathematische Probleme. Das Risiko von Halluzinationen wird minimiert. |
| **Programmierung** | 0.3 - 0.7 | 0.8 - 0.95 | Erzeugt funktionalen und korrekten Code, erlaubt dem Modell aber, elegante oder idiomatische Lösungswege zu finden. |

### Support und Programme

Programmfehler bitte in 'Issues' melden, Diskussionen und Fragen in 'Discussions'. Ausführbare Programme finden Sie im 'Releases'-Bereich.

***

![overview diagram](./gem-pro.png)

***

### Purpose

This program enables interaction with a Google Gemini AI model to send prompts and receive responses in various formats (Markdown, ANSI, HTML). It aims to seamlessly integrate AI capabilities into the user's workflow by supporting various input and output channels along with a history feature.

The program is available for BSD, Linux, macOS, and Windows.

### Google Gemini AI

This program allows the use of various models from the Google Gemini AI family. Many parameters of the AI model, such as the number of responses or variance, can be adjusted in the configuration.


### Google Gemini KI models

The Gemini AI family consists of the following models:

|  | Lite | Flash | Pro |
|---|---|---|---|
| Best for | high volume cost-efficient tasks | fast performance on everyday tasks | coding and highly complex tasks |
| Speed | high | medium | low |
| Performance | low | medium | high |
| Cost | low | medium | high |

### Features

* Multiple input sources (terminal, file polling, HTTP).
* Support for file uploads as part of the prompt.
* Configurable output formats (Markdown, ANSI, HTML).
* History storage for each output format.
* Chat mode for conversational interactions.
* Proxy support for environments without direct internet access.
* Detailed configuration options via YAML and command-line flags.
* Display of available Gemini models.
* OS-specific configurations (e.g., for external viewer applications).
* Embedding of default configuration files and web assets for easy deployment.
* Good user guidance through informative outputs (configuration, model info, usage text).
* Processing of metadata from the Gemini response (token usage, citations, grounding info like web sources).

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

Notes concerning the non-chat mode (default):

* Each prompt is treated independently.
* The AI does not remember previous interactions.
* Files are sent with every prompt.

Notes concerning the chat mode (-chatmode flag):

* The AI remembers the conversation history within a session.
* Files are sent only with the initial prompt.

Important: This application is intended for use with Google Gemini AI using a personal Gemini API key. Please note the terms of service and privacy information in the following section.

### Terms of Service and Privacy Notes

Refer to the [Google Terms of Service](https://policies.google.com/terms) and the [Generative AI Prohibited Use Policy](https://policies.google.com/terms/generative-ai/use-policy) for Gemini. For more information on privacy, please visit the [Gemini Apps Privacy Hub](https://support.google.com/gemini?p=privacy_help).

**Important**: The freely available version of Google Gemini AI uses your input and output data to improve the model. Therefore, do not process any private or confidential data. Different privacy policies apply to the paid version.

See also [Gemini Apps FAQ](https://gemini.google.com/faq?hl=en).

### Technical Notes

Each Gemini AI model has a maximum output token limit, which restricts the response length. One token roughly corresponds to approximately 4 characters, 100 tokens to about 60 to 80 English words. If, for example, the AI model defines the output token limit as 65536 tokens, that corresponds to a maximum response length of 39300 to 52400 English words.

### Required Prerequisites

A personal Gemini API key from Google is required for use. Configure the API key as an environment variable ```GEMINI_API_KEY```. An internet proxy can be optionally configured for the internet connection.

* Configure the API key in the program environment:
  * macOS, Linux: ```export GEMINI_API_KEY=Your-API-Key```
  * Windows: ```setx GEMINI_API_KEY Your-API-Key``` (requires terminal restart)

### File Handling

Depending on the use case, various options are available for utilizing file contents within the "Prompt" context.

* The content of a textual file is directly inserted into the prompt.
* A set of files (1 or more) is defined via the command line:
  * The files subsequently automatically become part of the prompt.
* A set of files (1 or more) is first uploaded to the Google File Store:
  * Via the '-include-files' option, the files subsequently become part of the prompt.
  
The aforementioned variants can be combined and do not differ in their effect on the prompt. If the set of files to be considered already exists in the Google File Store, it is sufficient to reference it in the prompt. Repeated, time-consuming uploading with every new prompt is therefore not required. The lifespan of the files in the Google File Store is limited. The files are automatically deleted after a certain time (e.g., 48 hours). The set of files in the Google File Store is considered as a 'unit' and is always fully referenced in the prompt.

### Handling an Explicit Cache
A cache contains 1-n files. The files are tokenized when the cache is created. A cache therefore offers two advantages:

* The cache can be referenced in the prompt.
* The complex tokenization process is only performed once.

The larger the cache, the more advantageous its use. A cache is chargeable. When used correctly, faster response times and lower total costs are possible. The lifespan of a cache can be configured (e.g., 4 hours). The chosen AI model must support 'Caching'. Upon creation, the cache is tokenized by the currently selected AI model. Subsequently, the cache can only be used by this AI model.

In a prompt, only one cache can be referenced. Therefore, all cache operations always act on a model-specific cache with the logical name "gem-pro-cache" (configurable). This is included in the prompt using the option '-include-cache'.

### Implicit Cache in Chat Mode

Chat mode offers an excellent way of (implicit) caching. Here, the files are only transferred and tokenized when a session is opened, i.e., as part of the first prompt. Gemini caches this data so that it is (implicitly) available for subsequent prompts.

### Tooling

To extend the capabilities of the AI model, specialized tools can be utilized. These enable the model to access external information or perform complex operations that go beyond mere text generation.

#### Google Search

When this tool is enabled, the AI model can access Google Search to find up-to-date information and verify facts. This improves the accuracy and relevance of the responses, particularly for questions about current events or specific knowledge.

#### Code Execution

This tool allows the AI model to generate, execute, and validate code (typically Python) to find a solution to a prompt. This is especially useful for mathematical calculations, data analysis, or logical problems that require a precise execution of steps. It enables the model to solve tasks that demand accurate and repeatable results.

### Gemini AI models

In the context of this application, the following AI model lines are of particular interest:

* **Flash** line with strengths in high speed and cost-efficiency, versatility
* **Pro** line with strengths in maximum performance and accuracy, advanced reasoning, programming support

Depending on the task, using both model lines can also be helpful:

* Flash line: Generation of multiple solution variants
* User: Selection of the best solution variant
* Pro line: Refinement of the selected solution variant

### Important Parameters

*   **Temperature**: Controls creativity. A low value (near 0.0) makes the output more deterministic and focused on the most likely words. A high value (above 0.7) increases randomness, encouraging more diverse and unexpected results.

*   **TopP (Nucleus Sampling)**: Filters word choices based on cumulative probability. A low value (e.g., 0.1) limits the selection to a very small pool of the most probable words, making the response focused. A high value (near 1.0) allows for a wider range of word choices, thus increasing diversity.

#### Default Values

The default values for Gemini models are typically:
*   **Temperature**: 1.0
*   **TopP**: 0.95

#### Recommended Values for Different Tasks

These values are recommendations. The ideal settings depend on the AI model and the specific task.

| Task | Temperature | TopP | Description |
| :--- | :--- | :--- | :--- |
| **General Tasks** | ~0.7 | ~0.95 | A good middle ground for balanced and coherent, yet not overly rigid, responses. |
| **Creative Tasks** | 0.7 - 1.0+ | 0.9 - 0.99 | Encourages novel ideas, brainstorming, and the creation of unique texts like poems or stories. |
| **Facts & Logic** | 0.0 - 0.4 | < 0.8 | Maximizes precision and factual accuracy for explanations or math problems. The risk of hallucination is minimized. |
| **Programming** | 0.3 - 0.7 | 0.8 - 0.95 | Generates functional and correct code, but allows the model to find elegant or idiomatic solutions. |

### Support and Programs

Please report program errors in 'Issues'. For discussions and questions, please use 'Discussions'. Executable programs are available in the 'Releases' section.
