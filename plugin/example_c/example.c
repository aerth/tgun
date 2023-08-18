#include <tgun.h>
#include <stdio.h>     /* for printf */
#include <getopt.h>
#include <string.h>

char *headerconfig = NULL;

int do_tgun(char *url){
  char* b = get_url_headers(url, headerconfig);
  if (!b) {
    printf("err: %s\n", tgunerr());
    return 1;
  }
  printf("%s\n", b);
  free(b);
  return 0;
}

char *usage = R"(
 _
| |_ __ _ _   _ _ __
| __/ _` | | | | '_ \
| || (_| | |_| | | | |
 \__\__, |\__,_|_| |_|
    |___/ https://github.com/aerth/tgun


example: 127.0.0.1:9050 socks5 proxy, custom UA and headers
  ./tgun -t --ua MyBrowser/1.0 --header "foo=bar;bar=foo" https://httpbin.org/get

flags:
  -h --help
  -t --tor (port 9150 or 9050 depending on platform)
  -p --proxy  eg: socks5://127.0.0.1:1080 ($PROXY env)
  --ua user-agent
  --headers   eg: foo=bar;bar=foo
)";

int main(int argc, char **argv){
  int c;
  while (1) {
    int option_index = 0;
    static struct option long_options[] = {
      {"help",     no_argument, 0,  0 },
      {"tor",     no_argument, 0,  0 },
      {"proxy",  required_argument,       0,  0 },
      {"ua",  required_argument,       0,  0 },
      {"headers",    required_argument, 0,  0 },
      {0,         0,                 0,  0 }
    };

    c = getopt_long(argc, argv, "htp:",
        long_options, &option_index);
    if (c == -1)
      break;

    switch (c) {
      case 0:
        switch (option_index){
          case 0:
            fprintf(stderr,"%s\n", usage);
            return 1;
          case 2:
            easy_proxy(optarg);
            goto Again;
          case 3:
            easy_ua(optarg);
            goto Again;
          case 4:
            headerconfig = malloc(1024);
            strncpy(headerconfig, optarg, 1024);
#ifdef DEBUG
            fprintf(stderr, "header config: %s\n", headerconfig);
#endif
            goto Again;
          default:
            printf("option %d:%s\n", option_index, long_options[option_index].name);
            if (optarg)
              printf(" with arg %s\n", optarg);
            printf("\n");
            goto Again;
        }
      case 'h':
        fprintf(stderr,"%s\n", usage);
        return 1;
      case 't':
        easy_proxy("tor");
        break;
      case 'p':
        easy_proxy(optarg);
        break;
      case '?':
        return 1;
      default:
        printf("?? getopt returned character code 0%o ??\n", c);
    }
Again:
    continue;
  }

  if (optind != argc - 1) {
    fprintf(stderr,"%s\n", usage);
    return 1;
  }
  char *url =  argv[optind++];
  return do_tgun(url);
}
