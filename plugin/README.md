# tgun

here we harness go networking in a c program.
at its most basic, fetching a url

compile with `gcc -o example_error example_c/example_error.c -ltgun`

```

#include <tgun.h>

// example usage
int using_tgun(char *url){
  char* b = get_url(url);
  if (!b) {
      // oops, theres an error string waiting
      printf("err: %s\n", tgunerr());
      return 1;
  }
  printf("%s\n", b); // print resp (its null terminated)
  free(b); // free when finished
  return 0;
}

```
