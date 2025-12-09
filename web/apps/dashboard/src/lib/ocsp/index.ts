import * as pvutils from 'pvutils'
import { OCSPRequest, OCSPResponse, BasicOCSPResponse, SingleResponse, ResponseBytes, ResponseData, CertID, CertStatus } from '@peculiar/asn1-ocsp'
import { Certificate, AlgorithmIdentifier } from '@peculiar/asn1-x509'
import { AsnParser, AsnSerializer, OctetString } from '@peculiar/asn1-schema'
import * as crypto from 'crypto'

export interface OCSPResponderConfig {
  caCertPem: string
  caKeyPem: string
}

export class OCSPResponder {
  private caCert: Certificate
  private caCertDer: ArrayBuffer
  private caPrivateKey: crypto.KeyObject

  constructor(config: OCSPResponderConfig) {
    // Parse CA certificate from PEM
    this.caCertDer = this.pemToDer(config.caCertPem, 'CERTIFICATE')
    this.caCert = AsnParser.parse(this.caCertDer, Certificate)

    // Parse CA private key from PEM
    this.caPrivateKey = crypto.createPrivateKey(config.caKeyPem)
  }

  private pemToDer(pem: string, _type: string): ArrayBuffer {
    const lines = pem.split('\n')
    const base64 = lines
      .filter(line => !line.startsWith(`-----`))
      .join('')
    return pvutils.stringToArrayBuffer(pvutils.fromBase64(base64))
  }

  async handleOCSPRequest(requestBytes: ArrayBuffer): Promise<ArrayBuffer> {
    try {
      // Parse the OCSP request
      const ocspRequest = AsnParser.parse(requestBytes, OCSPRequest)

      // Get the first request (typically only one)
      const reqList = ocspRequest.tbsRequest.requestList
      if (reqList.length === 0) {
        return this.buildErrorResponse(2) // malformedRequest
      }

      // Build response for each requested certificate
      const singleResponses: SingleResponse[] = []

      for (const req of reqList) {
        const singleResponse = this.buildSingleResponse(req.reqCert)
        singleResponses.push(singleResponse)
      }

      // Build the complete OCSP response
      return this.buildSuccessfulResponse(singleResponses)
    } catch (error) {
      console.error('Error processing OCSP request:', error)
      return this.buildErrorResponse(1) // internalError
    }
  }

  private buildSingleResponse(certId: CertID): SingleResponse {
    const now = new Date()
    const nextUpdate = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000) // 7 days

    const singleResponse = new SingleResponse({
      certID: certId,
      certStatus: new CertStatus({ good: null }), // Always return "good" status
      thisUpdate: now,
      nextUpdate: nextUpdate,
    })

    return singleResponse
  }

  private async buildSuccessfulResponse(singleResponses: SingleResponse[]): Promise<ArrayBuffer> {
    const now = new Date()

    // Build ResponseData
    const responseData = new ResponseData({
      version: 0,
      responderID: {
        byName: this.caCert.tbsCertificate.subject,
      },
      producedAt: now,
      responses: singleResponses,
    })

    // Serialize ResponseData for signing
    const responseDataDer = AsnSerializer.serialize(responseData)

    // Sign the response data with CA private key
    const signature = this.signData(responseDataDer)

    // Build BasicOCSPResponse
    const basicOCSPResponse = new BasicOCSPResponse({
      tbsResponseData: responseData,
      signatureAlgorithm: new AlgorithmIdentifier({
        algorithm: '1.2.840.10045.4.3.2', // ECDSA with SHA-256
      }),
      signature: signature, // ArrayBuffer containing the signature
      certs: [this.caCert], // Include the signing certificate
    })

    // Serialize BasicOCSPResponse
    const basicOCSPResponseDer = AsnSerializer.serialize(basicOCSPResponse)

    // Build ResponseBytes
    const responseBytes = new ResponseBytes({
      responseType: '1.3.6.1.5.5.7.48.1.1', // id-pkix-ocsp-basic
      response: new OctetString(basicOCSPResponseDer),
    })

    // Build final OCSPResponse
    const ocspResponse = new OCSPResponse({
      responseStatus: 0, // successful
      responseBytes: responseBytes,
    })

    return AsnSerializer.serialize(ocspResponse)
  }

  private buildErrorResponse(status: number): ArrayBuffer {
    const ocspResponse = new OCSPResponse({
      responseStatus: status,
    })
    return AsnSerializer.serialize(ocspResponse)
  }

  private signData(data: ArrayBuffer): ArrayBuffer {
    const sign = crypto.createSign('SHA256')
    sign.update(Buffer.from(data))
    const signature = sign.sign(this.caPrivateKey)
    return signature.buffer.slice(signature.byteOffset, signature.byteOffset + signature.byteLength)
  }
}

export async function createOCSPResponder(caCertPem: string, caKeyPem: string): Promise<OCSPResponder> {
  return new OCSPResponder({ caCertPem, caKeyPem })
}
