Istruzioni per eseguire il programma:
1) Inserire i path relativi ai file di testo da leggere nel file id configurazione "word_files.json"
2) Configurare il file "workers.json" scegliendo il numero di workers e le porte a cui il master dovrà contattarli per la rpc
3) Lanciare i relativi workers:
	3a) Utilizzando gli script bash workers.sh con argomenti la lista delle porte
	3b) Lanciando manualmente l'eseguibile './worker/worker' con argomento la porta da usare per l'rpc
4) Lanciare il master dall' eseguibile './master/master' :
	4a) Lanciando senza parametri si utilizzeranno i path di default per cercare i file di configurazioni (settati nel passo 1 e 2)
	4b) Lanciando con i parametri: [path file json parole] [path file json worker] se si ha problemi con il path o si vogliono utilizzare altri 			file di configurazione.
5) Da fare solo nel caso di aver scelto il punto 3a): Bisogna terminare gli worker. Lo script "kill_workers.sh" fa questo prendendo come parametro le porte dove sono i processi da terminare (Attenzione lo script uccide tutti i processi che utilizzino tale porta).

Informazioni programma:
Il programma è stato sviluppato e testato su una distro linux e quindi utilizza per i path il formato linux, inoltre il programma utilizza i path relativi.
Il risultato del wordcount viene stampato a video.

