import * as x509 from '@peculiar/x509'
import { Crypto } from '@peculiar/webcrypto'
import { OCSPRequest, OCSPResponse, BasicOCSPResponse, SingleResponse, ResponseBytes, ResponseData, CertStatus, ResponderID } from '@peculiar/asn1-ocsp'
import { AlgorithmIdentifier, Certificate } from '@peculiar/asn1-x509'
import { AsnParser, AsnSerializer, OctetString } from '@peculiar/asn1-schema'
import * as nodeCrypto from 'crypto'

// Set up the crypto provider for @peculiar/x509
const crypto = new Crypto()
x509.cryptoProvider.set(crypto as unknown as globalThis.Crypto)

export interface OCSPResponderConfig {
  caCertPem: string
  caKeyPem: string
}

export class OCSPResponder {
  private caCert: x509.X509Certificate
  private caPrivateKey: nodeCrypto.KeyObject

  constructor(config: OCSPResponderConfig) {
    // Parse CA certificate from PEM
    this.caCert = new x509.X509Certificate(config.caCertPem)

    // Parse CA private key from PEM
    this.caPrivateKey = nodeCrypto.createPrivateKey(config.caKeyPem)
  }

  async handleOCSPRequest(requestBytes: ArrayBuffer): Promise<ArrayBuffer> {
    try {
      // Parse the OCSP request
      const ocspRequest = AsnParser.parse(requestBytes, OCSPRequest)

      // Get the request list
      const reqList = ocspRequest.tbsRequest.requestList
      if (reqList.length === 0) {
        return this.buildErrorResponse(2) // malformedRequest
      }

      // Build response for each requested certificate
      const singleResponses: SingleResponse[] = []
      const now = new Date()
      const nextUpdate = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000) // 7 days

      for (const req of reqList) {
        // Create a new CertStatus with good status
        const certStatus = new CertStatus()
        certStatus.good = null

        const singleResponse = new SingleResponse()
        singleResponse.certID = req.reqCert
        singleResponse.certStatus = certStatus
        singleResponse.thisUpdate = now
        singleResponse.nextUpdate = nextUpdate

        singleResponses.push(singleResponse)
      }

      // Build the complete OCSP response
      return await this.buildSuccessfulResponse(singleResponses)
    } catch (error) {
      console.error('Error processing OCSP request:', error)
      return this.buildErrorResponse(1) // internalError
    }
  }

  private async buildSuccessfulResponse(singleResponses: SingleResponse[]): Promise<ArrayBuffer> {
    const now = new Date()

    // Build ResponderID using the CA certificate subject
    const caCertAsn = AsnParser.parse(this.caCert.rawData, Certificate)
    const responderID = new ResponderID()
    responderID.byName = caCertAsn.tbsCertificate.subject

    // Build ResponseData
    const responseData = new ResponseData()
    responseData.version = 0
    responseData.responderID = responderID
    responseData.producedAt = now
    responseData.responses = singleResponses

    // Serialize ResponseData for signing
    const responseDataDer = AsnSerializer.serialize(responseData)

    // Sign the response data with CA private key
    const signature = this.signData(responseDataDer)

    // Build BasicOCSPResponse
    const basicOCSPResponse = new BasicOCSPResponse()
    basicOCSPResponse.tbsResponseData = responseData
    basicOCSPResponse.signatureAlgorithm = new AlgorithmIdentifier({
      algorithm: '1.2.840.10045.4.3.2', // ECDSA with SHA-256
    })
    basicOCSPResponse.signature = signature
    basicOCSPResponse.certs = [caCertAsn]

    // Serialize BasicOCSPResponse
    const basicOCSPResponseDer = AsnSerializer.serialize(basicOCSPResponse)

    // Build ResponseBytes
    const responseBytes = new ResponseBytes()
    responseBytes.responseType = '1.3.6.1.5.5.7.48.1.1' // id-pkix-ocsp-basic
    responseBytes.response = new OctetString(basicOCSPResponseDer)

    // Build final OCSPResponse
    const ocspResponse = new OCSPResponse()
    ocspResponse.responseStatus = 0 // successful
    ocspResponse.responseBytes = responseBytes

    return AsnSerializer.serialize(ocspResponse)
  }

  private buildErrorResponse(status: number): ArrayBuffer {
    const ocspResponse = new OCSPResponse()
    ocspResponse.responseStatus = status
    return AsnSerializer.serialize(ocspResponse)
  }

  private signData(data: ArrayBuffer): ArrayBuffer {
    const sign = nodeCrypto.createSign('SHA256')
    sign.update(Buffer.from(data))
    const signature = sign.sign(this.caPrivateKey)
    return signature.buffer.slice(signature.byteOffset, signature.byteOffset + signature.byteLength)
  }
}

export async function createOCSPResponder(caCertPem: string, caKeyPem: string): Promise<OCSPResponder> {
  return new OCSPResponder({ caCertPem, caKeyPem })
}
