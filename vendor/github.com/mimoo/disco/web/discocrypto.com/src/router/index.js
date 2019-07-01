import Vue from 'vue'
import Router from 'vue-router'
//
import LandingPage from '@/components/LandingPage'
import GetStarted from '@/components/GetStarted'
// protocol
import protocolOverview from '@/components/handshakes/Overview'
import Keys from '@/components/Keys'
import Noise_K from '@/components/handshakes/Noise_K'
import Noise_N from '@/components/handshakes/Noise_N'
import Noise_X from '@/components/handshakes/Noise_X'
import Noise_NNpsk2 from '@/components/handshakes/Noise_NNpsk2'
import Noise_KK from '@/components/handshakes/Noise_KK'
import Noise_NK from '@/components/handshakes/Noise_NK'
import Noise_NX from '@/components/handshakes/Noise_NX'
import Noise_KX from '@/components/handshakes/Noise_KX'
import Noise_XK from '@/components/handshakes/Noise_XK'
import Noise_XX from '@/components/handshakes/Noise_XX'
import Noise_NNoob from '@/components/handshakes/Noise_NNoob'
// library
import libraryOverview from '@/components/library/Overview'
import Hashing from '@/components/library/Hashing'
import Encryption from '@/components/library/Encryption'
import DerivingKeys from '@/components/library/DerivingKeys'
import Signing from '@/components/library/Signing'
import IntegrityProtection from '@/components/library/IntegrityProtection'
import RandomNumbers from '@/components/library/RandomNumbers'
import PasswordHashing from '@/components/library/PasswordHashing'

Vue.use(Router)

export default new Router({
  scrollBehavior (to, from, savedPosition) {
    return { x: 0, y: 0 }
  },
  routes: [
    {
      path: '/',
      name: 'LandingPage',
      component: LandingPage
    },
    {
      path: '/get_started',
      name: 'GetStarted',
      component: GetStarted
    },
    // protocol
    {
      path: '/protocol/Overview',
      name: 'protocolOverview',
      component: protocolOverview
    },
    {
      path: '/protocol/Keys',
      name: 'Keys',
      component: Keys
    },
    {
      path: '/protocol/Noise_K',
      name: 'Noise_K',
      component: Noise_K
    },
    {
      path: '/protocol/Noise_N',
      name: 'Noise_N',
      component: Noise_N
    },
    {
      path: '/protocol/Noise_X',
      name: 'Noise_X',
      component: Noise_X
    },
    {
      path: '/protocol/Noise_NNpsk2',
      name: 'Noise_NNpsk2',
      component: Noise_NNpsk2
    },
    {
      path: '/protocol/Noise_KK',
      name: 'Noise_KK',
      component: Noise_KK
    },
    {
      path: '/protocol/Noise_NK',
      name: 'Noise_NK',
      component: Noise_NK
    },
    {
      path: '/protocol/Noise_NX',
      name: 'Noise_NX',
      component: Noise_NX
    },
    {
      path: '/protocol/Noise_KX',
      name: 'Noise_KX',
      component: Noise_KX
    },
    {
      path: '/protocol/Noise_XK',
      name: 'Noise_XK',
      component: Noise_XK
    },
    {
      path: '/protocol/Noise_XX',
      name: 'Noise_XX',
      component: Noise_XX
    },
    {
      path: '/protocol/Noise_NNoob',
      name: 'Noise_NNoob',
      component: Noise_NNoob
    }, 
    // library
    {
      path: '/library/Overview',
      name: 'libraryOverview',
      component: libraryOverview
    },
    {
      path: '/library/Hashing',
      name: 'Hashing',
      component: Hashing
    },
    {
      path: '/library/Encryption',
      name: 'Encryption',
      component: Encryption
    },
    {
      path: '/library/DerivingKeys',
      name: 'DerivingKeys',
      component: DerivingKeys
    },
    {
      path: '/library/Signing',
      name: 'Signing',
      component: Signing
    },
    {
      path: '/library/IntegrityProtection',
      name: 'IntegrityProtection',
      component: IntegrityProtection
    },
    {
      path: '/library/PasswordHashing',
      name: 'PasswordHashing',
      component: PasswordHashing
    },
    {
      path: '/library/RandomNumbers',
      name: 'RandomNumbers',
      component: RandomNumbers
    }
  ]
})
