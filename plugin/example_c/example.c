#include <getopt.h>
#include <string.h>
#include <tgun.h>
#include <stdio.h>     /* for printf */

char *headerconfig = NULL;


char *logotext = "  _\n"
                 " | |_ __ _ _   _ _ __\n"
                 " | __/ _` | | | | '_ \\\n"
                 " | || (_| | |_| | | | |\n"
                 "  \\__\\__, |\\__,_|_| |_|\n"
                 "     |___/ https://github.com/aerth/tgun\n"
                 "\n";

char *usage = "usage:\n"
              "   tgun [options] <url>\n"
              "\n"
              " flags:\n"
              "   -h --help\n"
              "   --version\n"
              "   -t --tor (auto socks5h port 9150 or 9050 depending on platform)\n"
              "   -o --output eg: \"/path/to/output.file\"\n"
              "   -p --proxy  eg: socks5h://127.0.0.1:1080 ($PROXY env)\n"
              "   -u --user-agent eg: 'MyBrowser/1.0'\n"
              "   -H --headers   eg: foo=bar;bar=foo\n"
              "   -c --timeout  (in milliseconds, try 1500)\n"
              "\n"
              " example (using 127.0.0.1:9050 socks5 proxy, a custom UA, wrap multiple headers in quotes)\n"
              "   tgun --tor --user-agent MyBrowser/1.0 --header \"foo=bar;bar=foo\" https://httpbin.org/get\n"
              " example (using short flags, timeout 1800 milliseconds and output to a file)\n"
              "   tgun -t -u 'MyBrowser/1.0' -H foo=bar -o /dev/tty -c 1800 https://httpbin.org/get\n";


// most of this is just setting up options
int main(int argc, char **argv){
    char *headerconfig = NULL;
    char *fileoutname = NULL;
    int c;
    static struct option long_options[] = {
        {"help",     no_argument, 0,  'h' }, // 0
        {"tor",     no_argument, 0,  't' }, // 1
        {"version", no_argument, 0, 'V' }, // 2
        {"proxy",  required_argument,       0,  'p' }, // 3
        {"user-agent",  required_argument,       0,  'u' }, // 3
        {"headers",    required_argument, 0,  'H' },// 4
        {"output", required_argument, 0, 'o' }, // 5
        {"timeout", required_argument, 0, 'c' }, // connect timeout
        {0,         0,                 0,  0 }
    };

    opterr=0;
    while (1) {
        int option_index = 0;
        c = getopt_long(argc, argv, "htvp:u:H:o:c:", long_options, &option_index);
        if (c == -1){
            break; // no args
        }
        switch (c) {
        case 'h':
            fprintf(stderr,"%s\n%s\nThanks for using tgun\n",logotext, usage);
            free(fileoutname);
            free(headerconfig);
            return 123;
        case 't':
            easy_proxy("tor");
            break;
        case 'V':
            fprintf(stdout, "%s\n", tgunversion());
            free(fileoutname);
            free(headerconfig);
            return 0;
        case 'p':
            easy_proxy(optarg);
            break;
        case 'u': // user-agent
            easy_ua(optarg);
            break;
        case 'H': // headers
            headerconfig = (char*) malloc(1024);
            strncpy(headerconfig, optarg, 1024);
            break;
        case 'o':
            fileoutname = (char*) malloc(1024);
            strncpy(fileoutname, optarg, 1024);
            if (strncmp(optarg, "-", 1024) == 0){
                // is already default. can use /dev/tty or /dev/{stderr|stdout} for real file
                free(fileoutname);
                fileoutname = NULL;
            }
            break;
        case 'c':
            if (!atoi(optarg)){
                fprintf(stderr, "weird timeout number: %s\n", optarg);
                return 142;
            }
            easy_timeout(atoi(optarg));
            break;
        case '?':
            free(fileoutname);
            free(headerconfig);
            return 4;
        default:
            printf("?? getopt returned character code 0%o ?? (%c) \n", c, c);
            free(fileoutname);
            free(headerconfig);
            return 3;
        }
    }

    if (optind != argc - 1) { // no url
        fprintf(stderr,"%s\n%s\n",logotext, usage);
        free(fileoutname);
        free(headerconfig);
        return 1;
    }

    // done parsing flags, now fetch url with tgun library
    char *url =  argv[optind++];
    int ret;
    ret = tgun_do("get", url, headerconfig, fileoutname);
    free(fileoutname);
    free(headerconfig);
    if (ret != 0) {
        fprintf(stderr, "Error: %s\n", tgunerr());
        return ret;
    }
    return 0;
}
