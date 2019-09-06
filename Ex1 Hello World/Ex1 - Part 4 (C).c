//https://www.onlinegdb.com/online_c_compiler

#include <pthread.h>	//Threads library
#include <stdio.h>

pthread_t thread_1;		//References to 2nd & 3st thread
pthread_t thread_2;

int i = 0;				//Global variable

void *inc() {
	for (int j=0; j<1000000; j++) {
		i += 1;
	}
}

void *dec() {
	for (int j=0; j<1000000; j++) {
		i -= 1;
	}
}

int main() {
	pthread_create(&thread_1, NULL, inc, NULL);	//2nd & 3st thread are executed, executing
	pthread_create(&thread_2, NULL, dec, NULL);	//functions inc/dec and passing no argument (NULL)
    
	pthread_join(thread_1, NULL);	//Wait for 2nd thread to finish
	pthread_join(thread_2, NULL);	//Wait for 3st thread to finish
	
	printf("Final value: %d\n", i);	//The final value is printed

	return 0;
}