/*
 * curl-like command
 *
 */

#include <tgun.h>
#include <stdio.h>     /* for printf */
#include <getopt.h>
#include <string.h>

char *headerconfig = NULL;


int do_tgun_mem(char *url){
    char* b = tgun_get_url_headers(url, headerconfig);
    if (!b) {
        printf("err: %s\n", tgunerr());
        return 1;
    }
    printf("%s", b);
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
  -o --output filename to output to (directory must exist)
  -p --proxy  eg: socks5://127.0.0.1:1080 ($PROXY env)
  -v --verbose 
  --ua user-agent
  --headers   eg: foo=bar;bar=foo
)";

int main(int argc, char **argv){
    char *headerconfig = NULL;
    char *fileoutname = NULL;
    int c;
    while (1) { // parse flags
        int option_index = 0;
        static struct option long_options[] = {
        {"help",     no_argument, 0,  0 },
        {"tor",     no_argument, 0,  0 },
        {"proxy",  required_argument,       0,  0 },
        {"ua",  required_argument,       0,  0 },
        {"headers",    required_argument, 0,  0 },
        {"output", required_argument, 0, 0 },
        {"verbose",   no_argument, 0,  0 },
        {0,         0,                 0,  0 }
    };

        c = getopt_long(argc, argv, "htpuhov:",
                        long_options, &option_index);
        if (c == -1)
            break;

        switch (c) {
            case 0:
                switch (option_index){
                    case 0: // --help
                        fprintf(stderr,"%s\n", usage);
                        return 1;
                    case 1: // --tor
                        tgun_easy_proxy("tor");
                        goto Again;
                    case 2: // --proxy
                        tgun_easy_proxy(optarg);
                        goto Again;
                    case 3: // -- ua
                        tgun_easy_ua(optarg);
                        goto Again;
                    case 4: // --headers
                        headerconfig = malloc(1024);
                        strncpy(headerconfig, optarg, 1024);
#ifdef DEBUG
                        fprintf(stderr, "header config: %s\n", headerconfig);
#endif
                        goto Again;
                    case 5: // --output
                        fileoutname = malloc(1024);
                        strncpy(fileoutname, optarg, 1024);
                        goto Again;
                    case 6: // --verbose
                        tgun_easy_verbose(1);
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
                tgun_easy_proxy("tor");
                break;
            case 'v':
                tgun_easy_verbose(1);
                break;
            case 'p':
                tgun_easy_proxy(optarg);
                printf("using proxy: %s\n", optarg);
                break;
            case 'o':
                fileoutname = malloc(1024);
                strncpy(fileoutname, optarg, 1024);
                break;
            case '?':
                return 1;
            default:
                printf("?? getopt returned character code 0%o ??\n", c);
                return 1;
        }
Again:
        continue;
    }

    if (optind != argc - 1) {
        fprintf(stderr,"%s\n", usage);
        return 1;
    }
    char *url =  argv[optind++];
    int ret;
    // do http request
    ret = tgun_do("get", url, headerconfig, fileoutname);
    if (ret != 0) {
        printf("err %d: %s",ret, tgunerr());
        tgunerr();
        return ret;
    }
    return 0;
}
