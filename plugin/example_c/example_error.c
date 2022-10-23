#include <tgun.h>

// example usage
int using_tgun(char *url){
  char* b = get_url(url);
  if (!b) {
      // oops, theres an error string waiting
      fprintf(stderr, "err: %s\n", tgunerr());
      return 1;
  }
  printf("%s\n", b); // print resp (its null terminated)
  free(b); // free when finished
  return 0;
}

int main(){
    // hopefully trigger an error
    return using_tgun("http://127.0.0.2:8080");
}
