// +build relic

#include "dkg_include.h"


#define N_max 250
#define N_bits_max 8  // log(250)  
#define T_max  ((N_max-1)/2)

// computes P(x) = a_0 + a_1*x + .. + a_n x^n (mod r)
// r being the order of G1
// writes P(x) in out and P(x).g2 in y if y is non NULL
// x being a small integer
void Zr_polynomialImage_export(byte* out, ep2_t y, const bn_t a, const int a_size, const byte x){
    bn_t image;
    bn_new(image);
    Zr_polynomialImage(image, y, a, a_size, x);
    // exports the result
    const int out_size = Fr_BYTES;
    bn_write_bin(out, out_size, image);
    bn_free(image);
}

// computes P(x) = a_0 + a_1*x + .. + a_n x^n (mod r)
// r being the order of G1
// writes P(x) in out and P(x).g2 in y if y is non NULL
// x being a small integer
void Zr_polynomialImage(bn_t image, ep2_t y, const bn_st *a, const int a_size, const byte x){
    bn_t r;
    bn_new(r); 
    g2_get_ord(r);

    // powers of x
    bn_t bn_x;         // maximum is |n|+|r| --> 264 bits
    ep_new(bn_x);
    bn_new_size(bn_x, BITS_TO_DIGITS(Fr_BITS+N_bits_max));
    bn_set_dig(bn_x, 1);
    
    // temp variables
    bn_t mult, acc;
    bn_new(mult);         // maximum --> 256+256 = 512 bits
    bn_new_size(mult, BITS_TO_DIGITS(2*Fr_BITS));
    bn_new(acc);         // maximum --> 512+1 = 513 bits
    bn_new_size(acc, BITS_TO_DIGITS(2*Fr_BITS+1));
    bn_set_dig(acc, 0);

    for (int i=0; i<a_size; i++) {
        bn_mul(mult, &a[i], bn_x);
        bn_add(acc, acc, mult);
        bn_mod_basic(acc, acc, r);
        // Use basic reduction as it's an 8-bits reduction 
        // in the worst case (|bn_x|<|r|+8 )
        bn_mul_dig(bn_x, bn_x, x);
        bn_mod_basic(bn_x, bn_x, r);
    }
    // copy the result
    bn_copy(image, acc); 

    // compute y = P(x).g2
    if (y) g2_mul_gen(y, acc);

    bn_free(acc)
    bn_free(mult);
    bn_free(r);
    bn_free(bn_x);
}

// computes Q(x) = A_0 + A_1*x + ... +  A_n*x^n  in G2
// and stores the point in y
// r is the order of G2
static void G2_polynomialImage(ep2_t y, const ep2_st* A, const int len_A,
         const byte x, const bn_t r){
    // powers of x
    bn_t bn_x;         // maximum is |n|+|r| --> 264 bits
    bn_new(bn_x);
    bn_new_size(bn_x, BITS_TO_DIGITS(Fr_BITS+N_bits_max));
    bn_set_dig(bn_x, 1);
    
    // temp variables
    ep2_t mult, acc;
    ep2_new(mult);         
    ep2_new(acc);
    ep2_set_infty(acc);

    for (int i=0; i < len_A; i++) {
        ep2_mul_lwnaf(mult, (ep2_st*)&A[i], bn_x);
        ep2_add_projc(acc, acc, mult);
        bn_mul_dig(bn_x, bn_x, x);
        // Use basic reduction as it's an 8-bits reduction 
        // in the worst case (|bn_x|<|r|+8 )
        bn_mod_basic(bn_x, bn_x, r);
    }
    // export the result
    ep2_copy(y, acc);
    ep2_norm(y, y);

    ep2_free(acc)
    ep2_free(mult);
    bn_free(bn_x);
}

// compute the nodes public keys from the verification vector
// y[i] = Q(i+1) for all nodes i, with:
// Q(x) = A_0 + A_1*x + ... +  A_n*x^n  in G2
void G2_polynomialImages(ep2_st *y, const int len_y, const ep2_st* A, const int len_A) {
    // order r
    bn_t r;
    bn_new(r); 
    g2_get_ord(r);
    for (byte i=0; i<len_y; i++) {
        //y[i] = Q(i+1)
        G2_polynomialImage(y+i , A, len_A, i+1, r);
    }
    bn_free(r);
}

// export an array of ep2_st into an array of bytes
// the length matching is supposed to be checked
void ep2_vector_write_bin(byte* out, const ep2_st* A, const int len) {
    const int size = (G2_BYTES/(SERIALIZATION+1));
    byte* p = out;
    for (int i=0; i<len; i++){
        ep2_write_bin_compact(p, &A[i], size);
        p += size;
    }
}

// imports an array of ep2_st from an array of bytes
// the length matching is supposed to be already done
int ep2_vector_read_bin(ep2_st* A, const byte* src, const int len){
    const int size = (G2_BYTES/(SERIALIZATION+1));
    byte* p = (byte*) src;
    for (int i=0; i<len; i++){
        if (ep2_read_bin_compact(&A[i], p, size) != RLC_OK)
            return RLC_ERR;
        p += size;
    }
    return RLC_OK;
}

// returns 1 if g2^x = y, where g2 is the generatot of G2
// returns 0 otherwise
int verifyshare(const bn_t x, const ep2_t y) {
    ep2_t res;
    ep2_new(res);
    g2_mul_gen(res, (bn_st*)x);
    return (ep2_cmp(res, (ep2_st*)y) == RLC_EQ);
}
